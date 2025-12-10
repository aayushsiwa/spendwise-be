package handlers

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ImportJSON(c *gin.Context) {
	var records []models.Record
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array"})
		return
	}

	// Update summary before importing
	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary before JSON import", "error", err)
	}

	importedCount := 0

	for _, rec := range records {
		if rec.Date == "" || rec.Description == "" || rec.Category == "" || rec.Type == "" {
			continue
		}

		// Normalize category
		category := strings.ToLower(strings.TrimSpace(rec.Category))
		dateStr := strings.TrimSpace(rec.Date)

		date, err := utils.ParseDate(dateStr)
		if err != nil {
			continue
		}

		var categoryID int
		err = h.DB.QueryRow(`SELECT id FROM categories WHERE name = ?`, category).Scan(&categoryID)
		if err != nil {
			res, err := h.DB.Exec(`INSERT INTO categories (name) VALUES (?)`, category)
			if err != nil {
				continue
			}
			lastID, _ := res.LastInsertId()
			categoryID = int(lastID)
		}

		// Get summary
		var currentBalance float64
		err = h.DB.QueryRow("SELECT closing_balance FROM summary WHERE month = ?", date[:7]).Scan(&currentBalance)
		if err == sql.ErrNoRows {
			currentBalance = 0
		} else if err != nil {
			continue
		}

		// Update balance based on record type
		if rec.Type == "income" {
			currentBalance += rec.Amount
		} else if rec.Type == "expense" {
			currentBalance -= rec.Amount
		}
		// For 'transfer', balance remains unchanged

		customID, err := h.GenerateCustomID(date)
		if err != nil {
			continue
		}

		_, err = h.DB.Exec(`INSERT INTO records (id, date, description, category_id, amount, type, note, balance) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			customID, date, rec.Description, categoryID, rec.Amount, rec.Type, rec.Note, currentBalance)
		if err != nil {
			log.Printf("Failed to insert record: %v", err)
			continue
		}

		// Update summary table
		if err := h.UpdateSummary(); err != nil {
			slog.Warn("Failed to update summary after record creation", "record_id", rec.ID, "error", err)
		}

		importedCount++
	}

	if err := h.RecalculateBalances(); err != nil {
		log.Printf("Failed to recalculate balances: %v", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":          "JSON import completed successfully",
		"records_imported": importedCount,
	})
}
