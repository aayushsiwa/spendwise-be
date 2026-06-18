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

	tx, err := h.DB.Begin()
	if err != nil {
		return errors.NewDatabase("Failed to begin transaction", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec("DELETE FROM summary"); err != nil {
		return errors.NewDatabase("Failed to clear summary", err)
	}
	if _, err = tx.Exec("DELETE FROM summary_details"); err != nil {
		return errors.NewDatabase("Failed to clear summary_details", err)
	}

	var minMonth sql.NullString
	if err = tx.QueryRow(`
		SELECT MIN(strftime('%Y-%m', date))
		FROM records
	`).Scan(&minMonth); err != nil {
		return errors.NewDatabase("Failed to get min month", err)
	}

	if !minMonth.Valid {
		slog.Info("No records found, summary will be empty")
		return tx.Commit()
	}

	maxMonth := time.Now().Format("2006-01")

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

	openingBalance := 0.0
	for m := minMonth.String; m <= maxMonth; m = nextMonth(m) {
		d := data[m]
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

	_, err = tx.Exec(`
		INSERT INTO summary_details ("ID", month, type, category_id, category_name, amount)
		SELECT
			r.id,
			strftime('%Y-%m', r.date) AS month,
			r.type,
			COALESCE(c.id, 0) AS category_id,
			COALESCE(c.name, 'uncategorized') AS category_name,
			SUM(r.amount) AS amount
		FROM records r
		LEFT JOIN categories c ON r.category_id = c.id
		WHERE r.type IN ('income', 'expense', 'transfer')
		GROUP BY month, r.type, COALESCE(c.id, 0), COALESCE(c.name, 'uncategorized')
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

// GetSummary handles GET /summary with optional query params: from, to, category, type.
// Computes totals and category breakdown directly from the records table.
func (h *Handler) GetSummary(c *gin.Context) {
	from := c.DefaultQuery("from", time.Now().Format("2006-01")+"-01")
	to := c.DefaultQuery("to", time.Now().Format("2006-01-02"))

	var totalIncome, totalExpense float64
	err := h.DB.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM records WHERE date >= ? AND date <= ?
	`, from, to).Scan(&totalIncome, &totalExpense)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to compute totals", err))
		return
	}

	var openingBalance float64
	err = h.DB.QueryRow(`
		SELECT COALESCE(SUM(
			CASE WHEN type = 'income' THEN amount
			     WHEN type = 'expense' THEN -amount
			     ELSE 0 END
		), 0) FROM records WHERE date < ?
	`, from).Scan(&openingBalance)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to compute opening balance", err))
		return
	}

	netInRange := totalIncome - totalExpense
	closingBalance := openingBalance + netInRange

	categoryFilter := c.Query("category")
	typeFilter := c.Query("type")

	detailQuery := `
		SELECT COALESCE(c.id, 0), COALESCE(c.name, ''), r.type, SUM(r.amount)
		FROM records r
		LEFT JOIN categories c ON r.category_id = c.id
		WHERE r.date >= ? AND r.date <= ?
	`
	detailArgs := []interface{}{from, to}

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

	detailQuery += " GROUP BY c.id, r.type ORDER BY r.type, SUM(r.amount) DESC"

	rows, err := h.DB.Query(detailQuery, detailArgs...)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to fetch category breakdown", err))
		return
	}
	defer rows.Close()

	incomes := make([]models.CategoryDetail, 0)
	expenses := make([]models.CategoryDetail, 0)
	for rows.Next() {
		var categoryID int
		var categoryName, recType string
		var amount float64
		if err := rows.Scan(&categoryID, &categoryName, &recType, &amount); err != nil {
			errors.HandleError(c, errors.NewDatabase("Failed to scan category detail", err))
			return
		}
		cd := models.CategoryDetail{CategoryID: categoryID, Category: categoryName, Amount: amount}
		if recType == "income" {
			incomes = append(incomes, cd)
		} else if recType == "expense" {
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
		"total_income", totalIncome,
		"total_expense", totalExpense,
	)

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

// nextMonth increments a YYYY-MM string by one month.
func nextMonth(m string) string {
	t, _ := time.Parse("2006-01", m)
	return t.AddDate(0, 1, 0).Format("2006-01")
}
