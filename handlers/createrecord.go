package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateRecord(c *gin.Context) {
	var rec models.Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		appErr := errors.NewInvalidInput("Invalid JSON body", err)
		errors.HandleError(c, appErr)
		return
	}

	// Validate record data
	validator := validation.NewValidator()
	validationErrs := validator.ValidateRecord(&rec)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Generate custom ID
	customId, err := h.GenerateCustomID(rec.Date)
	if err != nil {
		appErr := errors.NewInternal("Failed to generate record ID", err)
		errors.HandleError(c, appErr)
		return
	}

	rec.ID, err = strconv.Atoi(customId)
	if err != nil {
		appErr := errors.NewInternal("Failed to parse generated ID", err)
		errors.HandleError(c, appErr)
		return
	}

	// Get category ID
	var categoryId int
	err = h.DB.QueryRow("SELECT id FROM categories WHERE name = ?", strings.ToLower(rec.Category)).Scan(&categoryId)
	if err != nil {
		if err == sql.ErrNoRows {
			appErr := errors.NewInvalidInput("Category not found", err).WithDetails(map[string]interface{}{
				"category": rec.Category,
			})
			errors.HandleError(c, appErr)
		} else {
			appErr := errors.NewDatabase("Failed to find category", err)
			errors.HandleError(c, appErr)
		}
		return
	}

	// Get summary
	var currentBalance float64
	err = h.DB.QueryRow("SELECT closing_balance FROM summary WHERE month = ?", rec.Date[:7]).Scan(&currentBalance)
	if err == sql.ErrNoRows {
		currentBalance = 0
	} else if err != nil {
		appErr := errors.NewDatabase("Failed to get summary", err)
		errors.HandleError(c, appErr)
		return
	}

	// Update balance based on record type
	if rec.Type == "income" {
		currentBalance += rec.Amount
	} else if rec.Type == "expense" {
		currentBalance -= rec.Amount
	}
	// For 'transfer', balance remains unchanged

	// Insert record
	_, err = h.DB.Exec(`
		INSERT INTO records (id, date, description, category_id, amount, type, note, balance)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.Date, rec.Description, categoryId, rec.Amount, rec.Type, rec.Note, currentBalance)
	if err != nil {
		appErr := errors.NewDatabase("Failed to insert record", err)
		errors.HandleError(c, appErr)
		return
	}

	// Update summary
	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record creation", "record_id", rec.ID, "error", err)
	}

	slog.Info("Record created successfully", "record_id", rec.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Record with id %d created successfully", rec.ID),
		"id":      rec.ID,
	})
}
