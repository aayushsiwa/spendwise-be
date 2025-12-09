package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/secure"
	"aayushsiwa/expense-tracker/validation"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PatchRecord(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	var rec models.Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		appErr := errors.NewInvalidInput("Invalid JSON body", err)
		errors.HandleError(c, appErr)
		return
	}

	// Validate record data
	validationErrs = validator.ValidateRecord(&rec)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Encrypt sensitive fields with proper error handling
	if rec.Description != "" {
		encrypted, err := secure.Encrypt(rec.Description)
		if err != nil {
			appErr := errors.NewEncryption("Failed to encrypt description", err)
			errors.HandleError(c, appErr)
			return
		}
		rec.Description = encrypted
	}

	if rec.Note != "" {
		encrypted, err := secure.Encrypt(rec.Note)
		if err != nil {
			appErr := errors.NewEncryption("Failed to encrypt note", err)
			errors.HandleError(c, appErr)
			return
		}
		rec.Note = encrypted
	}

	// Get category ID
	var categoryId int
	err := h.DB.QueryRow("SELECT id FROM categories WHERE name = ?", rec.Category).Scan(&categoryId)
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

	// Check if record exists
	var exists int
	err = h.DB.QueryRow("SELECT COUNT(*) FROM records WHERE id = ?", id).Scan(&exists)
	if err != nil {
		appErr := errors.NewDatabase("Failed to check record existence", err)
		errors.HandleError(c, appErr)
		return
	}

	if exists == 0 {
		appErr := errors.NewNotFound(fmt.Sprintf("Record with ID %d not found", id), nil)
		errors.HandleError(c, appErr)
		return
	}

	// Update record
	_, err = h.DB.Exec(`
		UPDATE records 
		SET date = ?, description = ?, category_id = ?, amount = ?, type = ?, note = ?
		WHERE id = ?`,
		rec.Date, rec.Description, categoryId, rec.Amount, rec.Type, rec.Note, id)
	if err != nil {
		appErr := errors.NewDatabase("Failed to update record", err)
		errors.HandleError(c, appErr)
		return
	}

	// Update summary
	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record update", "record_id", id, "error", err)
	}

	rec.ID = id
	slog.Info("Record updated successfully", "record_id", rec.ID)
	c.JSON(http.StatusOK, rec)
}
