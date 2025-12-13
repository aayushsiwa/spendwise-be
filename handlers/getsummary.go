package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/utils"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateSummary updates the summary table and returns an error if it fails
func (h *Handler) UpdateSummary() (err error) {
	slog.Info("Updating summary...")

	tx, err := h.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return errors.NewDatabase("Failed to begin transaction", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Clear old data
	if _, err = tx.Exec("DELETE FROM summary"); err != nil {
		return errors.NewDatabase("Failed to clear summary", err)
	}
	if _, err = tx.Exec("DELETE FROM summary_details"); err != nil {
		return errors.NewDatabase("Failed to clear summary_details", err)
	}

	// Get min month from records
	var minMonth sql.NullString
	if err = tx.QueryRow(`
		SELECT MIN(strftime('%Y-%m', date))
		FROM records
	`).Scan(&minMonth); err != nil {
		return errors.NewDatabase("Failed to get min month", err)
	}

	if !minMonth.Valid {
		// No records at all → just exit cleanly
		slog.Info("No records found, summary will be empty")
		return tx.Commit()
	}

	// Use current month as maxMonth
	maxMonth := time.Now().Format("2006-01")

	// Get monthly aggregates from records
	rows, err := tx.Query(`
		SELECT
			strftime('%Y-%m', date) AS month,
			SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) AS total_income,
			SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) AS total_expense
		FROM records
		GROUP BY month
		ORDER BY month ASC
	`)
	if err != nil {
		return errors.NewDatabase("Failed to aggregate records", err)
	}
	defer rows.Close()

	type monthData struct {
		income  float64
		expense float64
	}
	data := make(map[string]monthData)

	for rows.Next() {
		var m string
		var inc, exp float64
		if err = rows.Scan(&m, &inc, &exp); err != nil {
			return errors.NewDatabase("Failed to scan summary row", err)
		}
		data[m] = monthData{inc, exp}
	}
	if err = rows.Err(); err != nil {
		return errors.NewDatabase("Error iterating summary rows", err)
	}

	// Iterate from minMonth to current month
	openingBalance := 0.0
	for m := minMonth.String; m <= maxMonth; m = utils.NextMonth(m) {
		d := data[m] // zero-value if not present
		net := d.income - d.expense
		closing := openingBalance + net

		_, err = tx.Exec(`
			INSERT INTO summary (month, total_income, total_expense, opening_balance, net_balance, closing_balance)
			VALUES (?, ?, ?, ?, ?, ?)
		`, m, d.income, d.expense, openingBalance, net, closing)
		if err != nil {
			return errors.NewDatabase("Failed to insert summary for month", err)
		}

		openingBalance = closing
	}

	// Populate summary_details for fast API
	_, err = tx.Exec(`
		INSERT INTO summary_details ("ID", month, type, category_id, category_name, amount)
		SELECT
			r.id,
			strftime('%Y-%m', r.date) AS month,
			r.type,
			c.id AS category_id,
			c.name AS category_name,
			SUM(r.amount) AS amount
		FROM records r
		JOIN categories c ON r.category_id = c.id
		WHERE r.type IN ('income', 'expense', 'transfer')
		GROUP BY month, r.type, c.id, c.name
	`)
	if err != nil {
		return errors.NewDatabase("Failed to insert summary details", err)
	}

	if err = tx.Commit(); err != nil {
		return errors.NewDatabase("Failed to commit summary transaction", err)
	}

	slog.Info("Summary updated successfully")
	return nil
}

// GetSummary handles GET /summary requests with filtering and pagination.
func (h *Handler) GetSummary(c *gin.Context) {
	queryParams := &models.QueryParams{
		PaginationFilterParams: models.PaginationFilterParams{
			Page:  1,
			Limit: 10,
		},
	}
	if err := c.ShouldBindQuery(queryParams); err != nil {
		appErr := errors.NewInvalidInput("Invalid query parameters", err)
		errors.HandleError(c, appErr)
		return
	}

	if err := h.UpdateSummary(); err != nil {
		errors.HandleError(c, err)
		return
	}

	if queryParams.From != "" || queryParams.To != "" {
		summary := getSummaryInDateRange(h, c, queryParams.From, queryParams.To)

		c.JSON(http.StatusOK, gin.H{"summary": &summary})
		return
	}

	fromMonth, toMonth := monthRangeFromQuery(queryParams)

	offset := (queryParams.Page - 1) * queryParams.Limit

	monthQuery := `
		SELECT month, total_income, total_expense, opening_balance, net_balance, closing_balance
		FROM summary
		WHERE month >= ? AND month <= ?
		ORDER BY month DESC
		LIMIT ? OFFSET ?
	`

	rows, err := h.DB.Query(monthQuery, fromMonth, toMonth, queryParams.Limit, offset)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to fetch summary", err))
		return
	}
	defer rows.Close()

	type monthRow struct {
		Month string
		models.Summary
	}
	months := []monthRow{}
	monthsList := []string{}
	for rows.Next() {
		var mr monthRow
		if err := rows.Scan(&mr.Month, &mr.TotalIncome, &mr.TotalExpense, &mr.Opening, &mr.Net, &mr.Closing); err != nil {
			errors.HandleError(c, errors.NewDatabase("Failed to scan summary row", err))
			return
		}
		months = append(months, mr)
		monthsList = append(monthsList, mr.Month)
	}
	if err := rows.Err(); err != nil {
		errors.HandleError(c, errors.NewDatabase("Error iterating summary rows", err))
		return
	}

	summaries := make(map[string]models.Summary)
	for _, m := range months {
		summaries[m.Month] = models.Summary{
			Expenses:     []models.CategoryDetail{},
			Incomes:      []models.CategoryDetail{},
			Net:          m.Net,
			Opening:      m.Opening,
			Closing:      m.Closing,
			TotalExpense: m.TotalExpense,
			TotalIncome:  m.TotalIncome,
		}
	}

	if len(monthsList) == 0 {
		c.JSON(http.StatusOK, gin.H{"summary": summaries})
		return
	}

	detailBase := strings.Builder{}
	args := []any{}

	whereClause, whereArgs := buildSummaryWhereClause(queryParams)
	detailBase.WriteString("month IN (")
	for i, m := range monthsList {
		if i > 0 {
			detailBase.WriteString(",")
		}
		detailBase.WriteString("?")
		args = append(args, m)
	}
	detailBase.WriteString(")")

	if strings.TrimSpace(whereClause) != "" {
		detailBase.WriteString(" AND (")
		detailBase.WriteString(whereClause)
		detailBase.WriteString(")")
		for _, a := range whereArgs {
			args = append(args, a)
		}
	}

	detailQuery := fmt.Sprintf(`
		SELECT "ID", month, type, category_id, category_name, SUM(amount) as total_amount
		FROM summary_details
		WHERE %s
		GROUP BY month, type, category_id, category_name
	`, detailBase.String())

	dRows, err := h.DB.Query(detailQuery, args...)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to fetch summary details", err))
		return
	}
	defer dRows.Close()

	for dRows.Next() {
		var month, recType string
		cd := models.CategoryDetail{}
		if err := dRows.Scan(&cd.ID, &month, &recType, &cd.CategoryID, &cd.Category, &cd.Amount); err != nil {
			errors.HandleError(c, errors.NewDatabase("Failed to parse summary detail row", err))
			return
		}

		s := summaries[month]
		if recType == "income" {
			s.Incomes = append(s.Incomes, cd)
		} else if recType == "transfer" {
			continue
		} else {
			s.Expenses = append(s.Expenses, cd)
		}
		summaries[month] = s
	}
	if err := dRows.Err(); err != nil {
		errors.HandleError(c, errors.NewDatabase("Error iterating summary detail rows", err))
		return
	}

	res := models.SummaryResponse{
		Summaries: summaries,
		PaginationMetadata: models.PaginationMetadata{
			Page:  queryParams.Page,
			Limit: queryParams.Limit,
		},
	}

	c.JSON(http.StatusOK, res)
}

// monthRangeFromQuery returns YYYY-MM formatted from/to bounds for queries.
// It respects QueryParams.From/To (if present) and TimeFrame params (timeframe/year/quarter/month).
// If nothing provided, returns very wide bounds (min..max) that will match your summary table.
func monthRangeFromQuery(q *models.QueryParams) (from string, to string) {
	// defaults (wide)
	from = "0000-01"
	to = time.Now().Format("2006-01")

	// explicit From/To take precedence if provided
	if q.From != "" {
		// expect q.From in "YYYY-MM" or "YYYY-MM-DD"; normalize to YYYY-MM
		if len(q.From) >= 7 {
			from = q.From[:7]
		}
	}
	if q.To != "" {
		if len(q.To) >= 7 {
			to = q.To[:7]
		}
	}

	// timeframe override (only if timeframe provided)
	if q.TimeFrame != "" {
		tf := strings.ToLower(q.TimeFrame)
		switch tf {
		case string(models.Year):
			year := q.Year
			if year != "" {
				from = fmt.Sprintf("%s-01", year)
				to = fmt.Sprintf("%s-12", year)
			}
		case string(models.Quarter):
			year := q.Year
			quarter := q.Quarter
			if year != "" && quarter != "" {
				switch quarter {
				case "1":
					from, to = fmt.Sprintf("%s-01", year), fmt.Sprintf("%s-03", year)
				case "2":
					from, to = fmt.Sprintf("%s-04", year), fmt.Sprintf("%s-06", year)
				case "3":
					from, to = fmt.Sprintf("%s-07", year), fmt.Sprintf("%s-09", year)
				case "4":
					from, to = fmt.Sprintf("%s-10", year), fmt.Sprintf("%s-12", year)
				}
			}
		case string(models.Month):
			year := q.Year
			month := q.Month
			if year != "" && month != "" {
				// ensure month is two digits
				if len(month) == 1 {
					month = "0" + month
				}
				from = fmt.Sprintf("%s-%s", year, month)
				to = from
			}
		}
	}

	return from, to
}

// BuildSummaryWhereClause builds an SQL fragment and args for filtering the summary_details table.
// NOTE: returns fragment WITHOUT a leading "WHERE". If it returns empty string, don't append filters.
func buildSummaryWhereClause(q *models.QueryParams) (string, []interface{}) {
	clauses := make([]string, 0, 6)
	args := make([]interface{}, 0, 6)

	// Type: income | expense | transfer
	if q.Type != "" {
		clauses = append(clauses, "type = ?")
		args = append(args, string(q.Type))
	}

	// Category: could be numeric id or name fragment
	if strings.TrimSpace(q.Category) != "" {
		c := strings.TrimSpace(q.Category)
		// if it's an integer, treat as category_id
		if id, err := strconv.Atoi(c); err == nil {
			clauses = append(clauses, "category_id = ?")
			args = append(args, id)
		} else {
			// match name case-insensitively using LIKE; using '%' wrapper for contains
			clauses = append(clauses, "LOWER(category_name) LIKE ?")
			args = append(args, "%"+strings.ToLower(c)+"%")
		}
	}

	// Amount filters
	// summary_details.amount holds aggregated amounts per category/month
	if q.MinAmount != 0 {
		clauses = append(clauses, "amount >= ?")
		args = append(args, q.MinAmount)
	}
	if q.MaxAmount != 0 {
		clauses = append(clauses, "amount <= ?")
		args = append(args, q.MaxAmount)
	}

	// Search: match category_name (summary_details doesn't have record description)
	if strings.TrimSpace(q.Search) != "" {
		s := strings.TrimSpace(q.Search)
		clauses = append(clauses, "LOWER(category_name) LIKE ?")
		args = append(args, "%"+strings.ToLower(s)+"%")
	}

	// If no clauses, return empty to indicate "no extra filters"
	if len(clauses) == 0 {
		return "", nil
	}

	return strings.Join(clauses, " AND "), args
}

func getSummaryInDateRange(h *Handler, c *gin.Context, from, to string) *models.Summary {
	query := `SELECT
  			COALESCE((SELECT SUM(CASE WHEN type='expense' THEN amount ELSE 0 END)
			FROM records
            WHERE date BETWEEN $1 AND $2), 0) AS total_expense,

  			COALESCE((SELECT SUM(CASE WHEN type='income' THEN amount ELSE 0 END)
			FROM records
			WHERE date BETWEEN $1 AND $2), 0) AS total_income,
  			
			COALESCE((SELECT SUM(
              CASE
                WHEN type = 'income' THEN amount
                WHEN type = 'expense' THEN -amount
                ELSE 0
              END)
            FROM records
            WHERE date < $1), 0) AS opening_balance,

			-- optional: net in range and closing
			COALESCE((SELECT SUM(
              CASE
                WHEN type = 'income' THEN amount
                WHEN type = 'expense' THEN -amount
                ELSE 0
              END)
            FROM records
            WHERE date BETWEEN $1 AND $2), 0) AS net_in_range,

			( -- closing = opening + net
			COALESCE((SELECT SUM(
              CASE
                WHEN type = 'income' THEN amount
                WHEN type = 'expense' THEN -amount
                ELSE 0
              END)
            FROM records
            WHERE date < $1), 0)
    		+
    		COALESCE((SELECT SUM(
              CASE
                WHEN type = 'income' THEN amount
                WHEN type = 'expense' THEN -amount
                ELSE 0
              END)
            FROM records
            WHERE date BETWEEN $1 AND $2), 0)
  			) AS closing_balance;`

	row, err := h.DB.Query(query, from, to)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to fetch summary from records", err))
		return nil
	}
	defer row.Close()
	var totalExpense, totalIncome, openingBalance, netInRange, closingBalance float64
	if row.Next() {
		if err := row.Scan(&totalExpense, &totalIncome, &openingBalance, &netInRange, &closingBalance); err != nil {
			errors.HandleError(c, errors.NewDatabase("Failed to scan summary from records", err))
			return nil
		}
	}

	summary := models.Summary{
		TotalExpense: totalExpense,
		TotalIncome:  totalIncome,
		Net:          netInRange,
		Opening:      openingBalance,
		Closing:      closingBalance,
	}

	if totalExpense == 0 && totalIncome == 0 {
		return &summary
	}

	// now we get type wise records from records table
	detailQuery := `
		SELECT 
			r.id, 
			strftime('%Y-%m', r.date) AS month,
			r.type,
			c.id AS category_id,
			c.name AS category_name,
			SUM(r.amount) AS amount
		FROM records r
		JOIN categories c ON r.category_id = c.id
		WHERE r.type IN ('income', 'expense')
		AND r.date >= $1
		AND r.date <= $2
		GROUP BY r.type, r.date;
		`

	dRows, err := h.DB.Query(detailQuery, from, to)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to fetch summary details", err))
		return nil
	}
	defer dRows.Close()

	for dRows.Next() {
		var month, recType string
		cd := models.CategoryDetail{}
		if err := dRows.Scan(&cd.ID, &month, &recType, &cd.CategoryID, &cd.Category, &cd.Amount); err != nil {
			errors.HandleError(c, errors.NewDatabase("Failed to parse summary detail row", err))
			return nil
		}

		if recType == "income" {
			summary.Incomes = append(summary.Incomes, cd)
		} else if recType == "transfer" {
			continue
		} else {
			summary.Expenses = append(summary.Expenses, cd)
		}
	}

	if err := dRows.Err(); err != nil {
		errors.HandleError(c, errors.NewDatabase("Error iterating summary detail rows", err))
		return nil
	}

	return &summary
}
