package handlers

import (
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/utils"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v4"
	"gorm.io/gorm"
)

func (h *Handler) ImportJSON(c *gin.Context) {
	ctx := c.Request.Context()
	var records []models.Record
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array"})
		return
	}

	tx := h.DB.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer func() {
		if tx.Error != nil {
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

		var categoryObj models.Category
		err = tx.Where("name = ?", category).First(&categoryObj).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				categoryObj = models.Category{
					ID:    shortuuid.New(),
					Name:  category,
					Icon:  "",
					Color: "",
				}
				if err := tx.Create(&categoryObj).Error; err != nil {
					continue
				}
			} else {
				continue
			}
		}

		var currentBalance float64
		err = tx.Model(&models.Record{}).Select("COALESCE(balance, 0)").Order("date DESC, id DESC").Limit(1).Scan(&currentBalance).Error
		if err != nil {
			currentBalance = 0
		}

		switch rec.Type {
		case "income":
			currentBalance += rec.Amount
		case "expense":
			currentBalance -= rec.Amount
		}

		customID, err := h.GenerateCustomID(date)
		if err != nil {
			continue
		}

		newRecord := models.Record{
			ID:          customID,
			Date:        date,
			Description: rec.Description,
			CategoryID:  &categoryObj.ID,
			Amount:      rec.Amount,
			Type:        rec.Type,
			Note:        rec.Note,
			Balance:     currentBalance,
		}

		if err := tx.Create(&newRecord).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to insert record during JSON import", "error", err)
			continue
		}

		importedCount++
	}

	if err := h.recalculateBalances(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to recalculate balances during JSON import", "error", err)
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Update summary
	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after JSON import", "error", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "JSON import completed successfully",
		"recordsImported": importedCount,
	})
}
