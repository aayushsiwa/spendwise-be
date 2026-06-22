package handlers

import (
	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) CreateRecord(c *gin.Context) {
	var rec models.Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		appErr := appErrors.NewInvalidInput("Invalid JSON body", err)
		appErrors.HandleError(c, appErr)
		return
	}

	// Validate record data
	validator := validation.NewValidator()
	validationErrs := validator.ValidateRecord(&rec)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Generate custom ID
	customId, err := h.GenerateCustomID(rec.Date)
	if err != nil {
		appErr := appErrors.NewInternal("Failed to generate record ID", err)
		appErrors.HandleError(c, appErr)
		return
	}

	rec.ID = customId

	// Get category ID
	var category models.Category
	err = h.DB.Select(`"ID"`).Where("name = ?", strings.ToLower(rec.Category)).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := appErrors.NewInvalidInput("Category not found", err).WithDetails(map[string]any{
				"category": rec.Category,
			})
			appErrors.HandleError(c, appErr)
		} else {
			appErr := appErrors.NewDatabase("Failed to find category", err)
			appErrors.HandleError(c, appErr)
		}
		return
	}

	rec.CategoryID = &category.ID

	// Compute running balance from actual records
	var currentBalance float64
	err = h.DB.Model(&models.Record{}).Select("COALESCE(balance, 0)").Order("date DESC, id DESC").Limit(1).Scan(&currentBalance).Error
	if err != nil {
		currentBalance = 0
	}

	switch rec.Type {
	case "income":
		currentBalance += rec.Amount
	case "expense":
		currentBalance -= rec.Amount
	}

	rec.Balance = currentBalance

	// Insert record
	if err := h.DB.Create(&rec).Error; err != nil {
		appErr := appErrors.NewDatabase("Failed to insert record", err)
		appErrors.HandleError(c, appErr)
		return
	}

	// Update summary
	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record creation", "record_id", rec.ID, "error", err)
	}

	slog.Info("Record created successfully", "record_id", rec.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Record with id %s created successfully", rec.ID),
		"ID":      rec.ID,
	})
}
