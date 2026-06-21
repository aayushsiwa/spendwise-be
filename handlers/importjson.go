package handlers

import (
	"log"
	"log/slog"
	"net/http"
	"strings"

	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/utils"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ImportJSON(c *gin.Context) {
	ctx := c.Request.Context()
	var records []models.Record
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array"})
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	importedCount := 0

	for _, rec := range records {
		if rec.Date == "" || rec.Description == "" || rec.Category == "" || rec.Type == "" {
			continue
		}

		category := strings.ToLower(strings.TrimSpace(rec.Category))
		dateStr := strings.TrimSpace(rec.Date)

		date, err := utils.ParseDate(dateStr)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to parse date", "date", dateStr, "error", err)
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

		var currentBalance float64
		err = h.DB.QueryRow("SELECT COALESCE(balance, 0) FROM records ORDER BY date DESC, id DESC LIMIT 1").Scan(&currentBalance)
		if err != nil {
			currentBalance = 0
		}

		if rec.Type == "income" {
			currentBalance += rec.Amount
		} else if rec.Type == "expense" {
			currentBalance -= rec.Amount
		}

		customID, err := h.GenerateCustomID(date)
		if err != nil {
			continue
		}

		_, err = h.DB.Exec(`INSERT INTO records (id, date, description, "categoryID", amount, type, note, balance) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			customID, date, rec.Description, categoryID, rec.Amount, rec.Type, rec.Note, currentBalance)
		if err != nil {
			log.Printf("Failed to insert record: %v", err)
			continue
		}

		importedCount++
	}

	if err := h.recalculateBalances(ctx, tx); err != nil {
		log.Printf("Failed to recalculate balances: %v", err)
	}

	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after JSON import", "error", err)
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "JSON import completed successfully",
		"recordsImported": importedCount,
	})
}
