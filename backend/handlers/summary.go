package handlers

import (
	"aayushsiwa/expense-tracker/db"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func UpdateSummary() {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer tx.Commit()

	// Clear existing summary
	_, err = tx.Exec("DELETE FROM summary")
	if err != nil {
		log.Printf("Failed to clear summary: %v", err)
		return
	}

	// Get all months in chronological order
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
		log.Printf("Failed to aggregate records: %v", err)
		return
	}
	defer rows.Close()

	var openingBalance float64 = 0

	for rows.Next() {
		var month string
		var income, expense, transfer float64

		if err := rows.Scan(&month, &income, &expense, &transfer); err != nil {
			log.Printf("Row scan error: %v", err)
			return
		}

		netBalance := income - expense + transfer
		closingBalance := openingBalance + netBalance

		_, err = tx.Exec(`
			INSERT INTO summary (month, total_income, total_expense, opening_balance, net_balance, closing_balance)
			VALUES (?, ?, ?, ?, ?, ?)
		`, month, income, expense, openingBalance, netBalance, closingBalance)

		if err != nil {
			log.Printf("Insert summary error for month %s: %v", month, err)
			return
		}

		openingBalance = closingBalance // carry forward for next month
	}
}

func GetSummary(c *gin.Context) {
	// Ensure the summary table is updated before fetching
	UpdateSummary()
	rows, err := db.DB.Query(`
	SELECT 
		month, total_income, total_expense,
		opening_balance, net_balance, closing_balance 
	FROM summary ORDER BY month DESC
`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch summary"})
		return
	}
	defer rows.Close()

	var summaries []map[string]interface{}

	for rows.Next() {
		var month string
		var income, expense, opening, net, closing float64
		err := rows.Scan(&month, &income, &expense, &opening, &net, &closing)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse summary row"})
			return
		}
		summaries = append(summaries, gin.H{
			"month":           month,
			"total_income":    income,
			"total_expense":   expense,
			"opening_balance": opening,
			"net_balance":     net,
			"closing_balance": closing,
		})
	}

	c.JSON(http.StatusOK, summaries)
}

func GetSummaryForFilter(c *gin.Context) {
	UpdateSummary()
	pathParts := strings.Split(c.Request.URL.Path, "/")
	if len(pathParts) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL path"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter type"})
	}
}

func GetSummaryByMonth(c *gin.Context, month string) {
	row := db.DB.QueryRow(`
	SELECT total_income, total_expense, opening_balance, net_balance, closing_balance
	FROM summary WHERE month = ?
`, month)

	var income, expense, openingBalance, netBalance, closingBalance float64
	err := row.Scan(&income, &expense, &openingBalance, &netBalance, &closingBalance)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No summary found for month"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"month":           month,
		"total_income":    income,
		"total_expense":   expense,
		"opening_balance": openingBalance,
		"net_balance":     netBalance,
		"closing_balance": closingBalance,
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
	if err != nil || (income == 0 && expense == 0) {
		c.JSON(http.StatusNotFound, gin.H{"error": "No data for category"})
		return
	}

	netBalance := income - expense

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
		return
	}

	var total float64

	err := db.DB.QueryRow(`
		SELECT SUM(amount) FROM records WHERE type = ?
	`, recordType).Scan(&total)
	if total == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No records found for type " + recordType})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total"})
		return
	}

	// If "type" is income, net is positive; if expense, net is negative
	netBalance := total
	if recordType == "expense" {
		netBalance = -total
	}

	c.JSON(http.StatusOK, gin.H{
		"type":            recordType,
		"total":           total,
		"net_balance":     netBalance,
		"closing_balance": netBalance, // similar logic as above
	})
}
