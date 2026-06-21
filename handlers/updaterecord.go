package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PatchRecord(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateRecordID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Check if record exists
	var exists int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM records WHERE id = ?", id).Scan(&exists)
	if err != nil {
		errors.HandleError(c, errors.NewDatabase("Failed to check record existence", err))
		return
	}
	if exists == 0 {
		errors.HandleError(c, errors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), nil))
		return
	}

	// Validate record data
	validationErrs = validator.ValidateRecord(&rec)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
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
		errors.HandleError(c, errors.NewDatabase("Failed to update record", err))
		return
	}

	// Recalculate all balances
	tx, err := h.DB.Begin()
	if err != nil {
		slog.Warn("Failed to start transaction for balance recalculation", "error", err)
	} else {
		if err := h.recalculateBalances(ctx, tx); err != nil {
			slog.Warn("Failed to recalculate balances after record update", "record_id", id, "error", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}

	// Update summary
	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record update", "record_id", id, "error", err)
	}

	slog.Info("Record updated successfully", "record_id", id)
	c.JSON(http.StatusOK, gin.H{"message": "Record updated", "ID": id})
}
