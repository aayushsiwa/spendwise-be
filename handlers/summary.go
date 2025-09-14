package handlers

import (
	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/utils"
	"aayushsiwa/expense-tracker/validation"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateSummary updates the summary table and returns an error if it fails
func UpdateSummary() (err error) {
	slog.Info("Updating summary...")

	tx, err := db.DB.Begin()
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
			SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) AS total_expense,
			SUM(CASE WHEN type = 'transfer' THEN amount ELSE 0 END) AS total_transfer
		FROM records
		GROUP BY month
		ORDER BY month ASC
	`)
	if err != nil {
		return errors.NewDatabase("Failed to aggregate records", err)
	}
	defer rows.Close()

	type monthData struct {
		income   float64
		expense  float64
		transfer float64
	}
	data := make(map[string]monthData)

	for rows.Next() {
		var m string
		var inc, exp, trf float64
		if err = rows.Scan(&m, &inc, &exp, &trf); err != nil {
			return errors.NewDatabase("Failed to scan summary row", err)
		}
		data[m] = monthData{inc, exp, trf}
	}
	if err = rows.Err(); err != nil {
		return errors.NewDatabase("Error iterating summary rows", err)
	}

	// Iterate from minMonth to current month
	openingBalance := 0.0
	for m := minMonth.String; m <= maxMonth; m = utils.NextMonth(m) {
		d := data[m] // zero-value if not present
		net := d.income - d.expense + d.transfer
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
		INSERT INTO summary_details (month, type, category_id, category_name, amount)
		SELECT
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

func GetSummary(c *gin.Context) {
	if err := UpdateSummary(); err != nil {
		errors.HandleError(c, err)
		return
	}

	// Step 1: Fetch monthly totals
	rows, err := db.DB.Query(`
        SELECT month, total_income, total_expense, opening_balance, net_balance, closing_balance 
        FROM summary ORDER BY month DESC
    `)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to fetch summary", err))
		return
	}
	defer rows.Close()

	summaries := make(map[string]models.MonthlySummary)

	for rows.Next() {
		var month string
		var income, expense, opening, net, closing float64
		if err := rows.Scan(&month, &income, &expense, &opening, &net, &closing); err != nil {
			errors.HandleError(c, errors.NewDatabase("Failed to parse summary row", err))
			return
		}
		summaries[month] = models.MonthlySummary{
			Expenses:     []models.CategoryDetail{},
			Incomes:      []models.CategoryDetail{},
			Net:          net,
			Opening:      opening,
			Closing:      closing,
			TotalExpense: expense,
			TotalIncome:  income,
		}
	}

	// Step 2: Fetch details
	detailRows, err := db.DB.Query(`
    SELECT 
        month, 
        type, 
        category_id, 
        category_name, 
        SUM(amount) as total_amount
    FROM summary_details
    GROUP BY month, type, category_id, category_name
	`)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to fetch summary details", err))
		return
	}
	defer detailRows.Close()

	for detailRows.Next() {
		var month, recordType, categoryName string
		var categoryID int
		var totalAmount float64

		if err := detailRows.Scan(&month, &recordType, &categoryID, &categoryName, &totalAmount); err != nil {
			errors.HandleError(c, errors.NewDatabase("Failed to parse summary detail row", err))
			return
		}

		summary := summaries[month]
		detail := models.CategoryDetail{
			CategoryID: categoryID,
			Category:   categoryName,
			Amount:     totalAmount,
		}

		if recordType == "income" {
			summary.Incomes = append(summary.Incomes, detail)
		} else {
			summary.Expenses = append(summary.Expenses, detail)
		}
		summaries[month] = summary
	}

	c.JSON(http.StatusOK, gin.H{"summary": summaries})
}

func GetSummaryForFilters(c *gin.Context) {
	// query params
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	month := c.Query("month")
	category := c.Query("category")
	recordType := c.Query("type")
	groupBy := c.Query("group_by")

	validator := validation.NewValidator()
	var validationErrs errors.ValidationErrors

	// Ensure the summary table is updated before fetching
	if err := UpdateSummary(); err != nil {
		errors.HandleError(c, err)
		return
	}

	var filters []string
	var args []interface{}

	if startDate != "" {
		startDate, err := utils.ParseDate(startDate)
		if err != nil {
			validationErrs = append(validationErrs, errors.NewValidationError("start_date", "Start date must be in YYYY-MM-DD format", startDate))
		}
		filters = append(filters, "r.date >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		endDate, err := utils.ParseDate(endDate)
		if err != nil {
			validationErrs = append(validationErrs, errors.NewValidationError("end_date", "End date must be in YYYY-MM-DD format", endDate))
			errors.HandleValidationErrors(c, validationErrs)
			return
		}
		filters = append(filters, "r.date <= ?")
		args = append(args, endDate)
	}
	if month != "" {
		startDate := month + "-01" // e.g., "2025-08-01"
		endDate := month + "-31"   // works fine, SQLite will handle shorter months
		filters = append(filters, "r.date BETWEEN ? AND ?")
		args = append(args, startDate, endDate)
	}
	if category != "" {
		filters = append(filters, "c.name = ?")
		args = append(args, category)
	}
	if recordType != "" {
		filters = append(filters, "r.type = ?")
		args = append(args, recordType)
	}
	allowedGroups := map[string]string{
		"month":    "strftime('%Y-%m', r.date)",
		"category": "c.name",
		"type":     "r.type",
	}

	groupCol, ok := allowedGroups[groupBy]
	if !ok || groupBy == "" {
		groupBy = "month"
		groupCol = allowedGroups["month"]
	}

	validationErrs = append(validationErrs, validator.GetErrors()...)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	query := fmt.Sprintf(`
    SELECT 
        %s AS %s,
        COALESCE(SUM(amount), 0) AS amount,
        c.name AS category
    FROM records r
    JOIN categories c ON r.category_id = c.id
	`, groupCol, groupBy)

	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}

	query += " GROUP BY category_id, " + groupBy

	slog.Debug("Executing query", "query", query, "args", args)

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		appErr := errors.NewDatabase("Failed to retrieve records", err)
		errors.HandleError(c, appErr)
		return
	}
	defer rows.Close()

	var summaries []map[string]interface{}

	for rows.Next() {
		var amount float64
		var group_by string
		var category string
		err := rows.Scan(&group_by, &amount, &category)

		if err != nil {
			appErr := errors.NewDatabase("Failed to parse summary row", err)
			errors.HandleError(c, appErr)
			return
		}

		summaries = append(summaries, gin.H{
			groupBy:    group_by,
			"category": category,
			"amount":   amount,
		})
	}

	if err = rows.Err(); err != nil {
		appErr := errors.NewDatabase("Error iterating through summary rows", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Summary retrieved successfully", "count", len(summaries))
	c.JSON(http.StatusOK, summaries)
}

func GetSummaryForFilter(c *gin.Context) {
	if err := UpdateSummary(); err != nil {
		errors.HandleError(c, err)
		return
	}

	pathParts := strings.Split(c.Request.URL.Path, "/")
	if len(pathParts) < 2 {
		appErr := errors.NewInvalidInput("Invalid URL path", nil)
		errors.HandleError(c, appErr)
		return
	}
	filterType := pathParts[len(pathParts)-2]
	value := pathParts[len(pathParts)-1]

	switch filterType {
	case "month":
		GetSummaryByMonth(c, value)
	case "category":
		GetSummaryByCategory(c, value)
	case "type":
		GetSummaryByType(c, value)
	default:
		appErr := errors.NewInvalidInput("Invalid filter type", nil).WithDetails(map[string]interface{}{
			"filter_type":   filterType,
			"allowed_types": []string{"month", "category", "type"},
		})
		errors.HandleError(c, appErr)
	}
}

func GetSummaryByMonth(c *gin.Context, month string) {
	// Get overall monthly totals
	row := db.DB.QueryRow(`
		SELECT 
			total_income, 
			total_expense, 
			opening_balance, 
			net_balance, 
			closing_balance
		FROM summary
		WHERE month = ?
	`, month)

	var totalIncome, totalExpense, openingBalance, netBalance, closingBalance float64
	if err := row.Scan(&totalIncome, &totalExpense, &openingBalance, &netBalance, &closingBalance); err != nil {
		appErr := errors.NewNotFound("No summary found for month", err).WithDetails(map[string]interface{}{
			"month": month,
		})
		errors.HandleError(c, appErr)
		return
	}

	// Get category breakdowns for that month
	rows, err := db.DB.Query(`
		SELECT 
			type, 
			category_id, 
			category_name, 
			SUM(amount) AS amount
		FROM summary_details
		WHERE month = ?
		GROUP BY type, category_id, category_name
	`, month)
	if err != nil {
		appErr := errors.NewDatabase("Failed to get summary details", err)
		errors.HandleError(c, appErr)
		return
	}
	defer rows.Close()

	var incomes []models.CategoryDetail
	var expenses []models.CategoryDetail

	for rows.Next() {
		var recordType, categoryName string
		var categoryID int
		var amount float64

		if err := rows.Scan(&recordType, &categoryID, &categoryName, &amount); err != nil {
			appErr := errors.NewDatabase("Failed to parse summary detail row", err)
			errors.HandleError(c, appErr)
			return
		}

		detail := models.CategoryDetail{
			CategoryID: categoryID,
			Category:   categoryName,
			Amount:     amount,
		}

		if recordType == "income" {
			incomes = append(incomes, detail)
		} else if recordType == "expense" {
			expenses = append(expenses, detail)
		}
	}

	slog.Info("Monthly summary retrieved", "month", month)
	c.JSON(http.StatusOK, gin.H{
		"summary": models.MonthlySummary{
			Expenses:     expenses,
			Incomes:      incomes,
			Net:          netBalance,
			Opening:      openingBalance,
			Closing:      closingBalance,
			TotalExpense: totalExpense,
			TotalIncome:  totalIncome,
		},
	})
}

func GetSummaryByCategory(c *gin.Context, category string) {
	row := db.DB.QueryRow(`
		SELECT 
			SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) AS income,
			SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) AS expense
		FROM records
		JOIN categories ON records.category_id = categories.id
		WHERE categories.name = ?
	`, category)

	var income, expense float64
	err := row.Scan(&income, &expense)
	if err != nil {
		appErr := errors.NewDatabase("Failed to get category summary", err)
		errors.HandleError(c, appErr)
		return
	}

	if income == 0 && expense == 0 {
		appErr := errors.NewNotFound("No data found for category", nil).WithDetails(map[string]interface{}{
			"category": category,
		})
		errors.HandleError(c, appErr)
		return
	}

	netBalance := income - expense

	slog.Info("Category summary retrieved", "category", category)
	c.JSON(http.StatusOK, gin.H{
		"category":        category,
		"total_income":    income,
		"total_expense":   expense,
		"net_balance":     netBalance,
		"closing_balance": netBalance, // assuming no opening balance for category-level
	})
}

func GetSummaryByType(c *gin.Context, recordType string) {
	if recordType != "income" && recordType != "expense" && recordType != "transfer" {
		appErr := errors.NewInvalidInput("Invalid record type", nil).WithDetails(map[string]interface{}{
			"type":          recordType,
			"allowed_types": []string{"income", "expense", "transfer"},
		})
		errors.HandleError(c, appErr)
		return
	}

	var total float64

	err := db.DB.QueryRow(`
		SELECT SUM(amount) FROM records WHERE type = ?
	`, recordType).Scan(&total)

	if err != nil {
		appErr := errors.NewDatabase("Failed to get type summary", err)
		errors.HandleError(c, appErr)
		return
	}

	if total == 0 {
		appErr := errors.NewNotFound("No records found for type", nil).WithDetails(map[string]interface{}{
			"type": recordType,
		})
		errors.HandleError(c, appErr)
		return
	}

	// If "type" is income, net is positive; if expense, net is negative
	netBalance := total
	if recordType == "expense" {
		netBalance = -total
	}

	slog.Info("Type summary retrieved", "type", recordType, "total", total)
	c.JSON(http.StatusOK, gin.H{
		"type":            recordType,
		"total":           total,
		"net_balance":     netBalance,
		"closing_balance": netBalance, // similar logic as above
	})
}
