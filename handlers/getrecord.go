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

func (h *Handler) GetRecord(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	row := h.DB.QueryRow(`
		SELECT r.id, r.date, r.description, c.name as category, r.amount, r.type, r.note
		FROM records r
		JOIN categories c ON r.category_id = c.id
		WHERE r.id = ?
	`, id)

	var rec models.Record
	if err := row.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Note); err != nil {
		if err == sql.ErrNoRows {
			appErr := errors.NewNotFound(fmt.Sprintf("Record with ID %d not found", id), err)
			errors.HandleError(c, appErr)
		} else {
			appErr := errors.NewDatabase("Failed to read record", err)
			errors.HandleError(c, appErr)
		}
		return
	}

	// Decrypt sensitive fields with proper error handling
	if rec.Description != "" {
		decrypted, err := secure.Decrypt(rec.Description)
		if err != nil {
			slog.Warn("Failed to decrypt description", "record_id", rec.ID, "error", err)
			rec.Description = "[Encryption Error]"
		} else {
			rec.Description = decrypted
		}
	}

	if rec.Note != "" {
		decrypted, err := secure.Decrypt(rec.Note)
		if err != nil {
			slog.Warn("Failed to decrypt Note", "record_id", rec.ID, "error", err)
			rec.Note = "[Encryption Error]"
		} else {
			rec.Note = decrypted
		}
	}

	slog.Info("Record retrieved successfully", "record_id", rec.ID)
	c.JSON(http.StatusOK, rec)
}
