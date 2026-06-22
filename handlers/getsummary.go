package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

// UpdateSummary refreshes the precomputed summary and summary_details tables.
// Called by mutation handlers (create/update/delete record) to keep aggregates current.
func (h *Handler) UpdateSummary() (err error) {
	slog.Info("Updating summary...")

	tx := h.DB.Begin()
	if tx.Error != nil {
		return errors.NewDatabase("Failed to begin transaction", tx.Error)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if err = tx.Exec("DELETE FROM summary").Error; err != nil {
		return errors.NewDatabase("Failed to clear summary", err)
	}
	if err = tx.Exec("DELETE FROM summary_details").Error; err != nil {
		return errors.NewDatabase("Failed to clear summary_details", err)
	}

	var minMonth sql.NullString
	if err = tx.Raw(`
		SELECT MIN(SUBSTR(date, 1, 7))
		FROM records
	`).Scan(&minMonth).Error; err != nil {
		return errors.NewDatabase("Failed to get min month", err)
	}

	if !minMonth.Valid {
		slog.Info("No records found, summary will be empty")
		return tx.Commit().Error
	}

	maxMonth := time.Now().Format("2006-01")

	rows, err := tx.Raw(`
		SELECT
			SUBSTR(date, 1, 7) AS month,
			SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) AS "totalIncome",
			SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) AS "totalExpense"
		FROM records
		GROUP BY month
		ORDER BY month ASC
	`).Rows()
	if err != nil {
		return errors.NewDatabase("Failed to aggregate records", err)
	}
	defer func() { _ = rows.Close() }()

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

	openingBalance := 0.0
	for m := minMonth.String; m <= maxMonth; m = nextMonth(m) {
		d := data[m]
		net := d.income - d.expense
		closing := openingBalance + net

		err = tx.Exec(`
			INSERT INTO summary (month, "totalIncome", "totalExpense", "openingBalance", "netBalance", "closingBalance")
			VALUES (?, ?, ?, ?, ?, ?)
		`, m, d.income, d.expense, openingBalance, net, closing).Error
		if err != nil {
			return errors.NewDatabase("Failed to insert summary for month", err)
		}

		openingBalance = closing
	}

	err = tx.Exec(`
		INSERT INTO summary_details ("ID", month, type, "categoryID", "categoryName", amount)
		SELECT
			r.id,
			SUBSTR(r.date, 1, 7) AS month,
			r.type,
			COALESCE(c."ID", '') AS "categoryID",
			COALESCE(c.name, 'uncategorized') AS "categoryName",
			SUM(r.amount) AS amount
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c."ID"
		WHERE r.type IN ('income', 'expense', 'transfer')
		GROUP BY r.id, month, r.type, COALESCE(c."ID", ''), COALESCE(c.name, 'uncategorized')
	`).Error
	if err != nil {
		return errors.NewDatabase("Failed to insert summary details", err)
	}

	if err = tx.Commit().Error; err != nil {
		return errors.NewDatabase("Failed to commit summary transaction", err)
	}

	slog.Info("Summary updated successfully")
	return nil
}

// GetSummary handles GET /summary with optional query params: from, to, category, type.
// Computes totals and category breakdown directly from the records table.
func (h *Handler) GetSummary(c *gin.Context) {
	from := c.DefaultQuery("from", time.Now().Format("2006-01")+"-01")
	to := c.DefaultQuery("to", time.Now().Format("2006-01-02"))

	var totalIncome, totalExpense float64
	err := h.DB.Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM records WHERE date >= ? AND date <= ?
	`, from, to).Row().Scan(&totalIncome, &totalExpense)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to compute totals", err))
		return
	}

	var openingBalance float64
	err = h.DB.Raw(`
		SELECT COALESCE(SUM(
			CASE WHEN type = 'income' THEN amount
			     WHEN type = 'expense' THEN -amount
			     ELSE 0 END
		), 0) FROM records WHERE date < ?
	`, from).Row().Scan(&openingBalance)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to compute opening balance", err))
		return
	}

	netInRange := totalIncome - totalExpense
	closingBalance := openingBalance + netInRange

	categoryFilter := c.Query("category")
	typeFilter := c.Query("type")

	detailQuery := `
		SELECT COALESCE(c.id, ''), COALESCE(c.name, ''), r.type, SUM(r.amount)
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
		WHERE r.date >= ? AND r.date <= ?
	`
	detailArgs := []any{from, to}

	var conditions []string
	if categoryFilter != "" {
		conditions = append(conditions, "LOWER(c.name) = ?")
		detailArgs = append(detailArgs, strings.ToLower(categoryFilter))
	}
	if typeFilter != "" {
		conditions = append(conditions, "r.type = ?")
		detailArgs = append(detailArgs, typeFilter)
	}
	if len(conditions) > 0 {
		detailQuery += " AND " + strings.Join(conditions, " AND ")
	}

	detailQuery += " GROUP BY c.id, c.name, r.type ORDER BY r.type, SUM(r.amount) DESC"

	rows, err := h.DB.Raw(detailQuery, detailArgs...).Rows()
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to fetch category breakdown", err))
		return
	}
	defer func() { _ = rows.Close() }()

	incomes := make([]models.CategoryDetail, 0)
	expenses := make([]models.CategoryDetail, 0)
	for rows.Next() {
		var categoryID string
		var categoryName, recType string
		var amount float64
		if err := rows.Scan(&categoryID, &categoryName, &recType, &amount); err != nil {
			errors.HandleError(c, errors.NewDatabase("Failed to scan category detail", err))
			return
		}
		cd := models.CategoryDetail{CategoryID: categoryID, Category: categoryName, Amount: amount}
		switch recType {
		case "income":
			incomes = append(incomes, cd)
		case "expense":
			expenses = append(expenses, cd)
		}
	}
	if err := rows.Err(); err != nil {
		errors.HandleError(c, errors.NewDatabase("Error iterating category details", err))
		return
	}

	summary := models.Summary{
		Expenses:     expenses,
		Incomes:      incomes,
		Net:          netInRange,
		Opening:      openingBalance,
		Closing:      closingBalance,
		TotalExpense: totalExpense,
		TotalIncome:  totalIncome,
	}

	slog.Info("Summary retrieved successfully",
		"from", from, "to", to,
		"totalIncome", totalIncome,
		"totalExpense", totalExpense,
	)

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

// nextMonth increments a YYYY-MM string by one month.
// (same implementation as before)
func nextMonth(m string) string {
	t, _ := time.Parse("2006-01", m)
	return t.AddDate(0, 1, 0).Format("2006-01")
}
