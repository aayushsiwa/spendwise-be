package handlers

import (
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

	var req models.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.NewInvalidInput("Invalid JSON body", err))
		return
	}

	validationErrs = validator.ValidatePatchRecord(&req)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	var setClauses []string
	var args []any

	if req.Date != nil {
		setClauses = append(setClauses, "date = ?")
		args = append(args, *req.Date)
	}
	if req.Description != nil {
		setClauses = append(setClauses, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Amount != nil {
		setClauses = append(setClauses, "amount = ?")
		args = append(args, *req.Amount)
	}
	if req.Type != nil {
		setClauses = append(setClauses, "type = ?")
		args = append(args, *req.Type)
	}
	if req.Note != nil {
		setClauses = append(setClauses, "note = ?")
		args = append(args, *req.Note)
	}
	if req.Category != nil {
		var categoryID int
		err := h.DB.QueryRow("SELECT id FROM categories WHERE name = ?", *req.Category).Scan(&categoryID)
		if err != nil {
			if err == sql.ErrNoRows {
				appErr := errors.NewInvalidInput("Category not found", err).WithDetails(map[string]any{
					"category": *req.Category,
				})
				errors.HandleError(c, appErr)
			} else {
				errors.HandleError(c, errors.NewDatabase("Failed to find category", err))
			}
			return
		}
		setClauses = append(setClauses, `"categoryID" = ?`)
		args = append(args, categoryID)
	}

	if len(setClauses) == 0 {
		slog.Info("No fields to update, returning success (idempotent)", "record_id", id)
		c.JSON(http.StatusOK, gin.H{"message": "No fields to update", "ID": id})
		return
	}

	query := fmt.Sprintf("UPDATE records SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)

	_, err = h.DB.Exec(query, args...)
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
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.Error("Failed to rollback transaction", "error", rbErr)
			}
		} else {
			if cErr := tx.Commit(); cErr != nil {
				slog.Error("Failed to commit transaction", "error", cErr)
			}
		}
	}

	// Update summary
	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record update", "record_id", id, "error", err)
	}

	slog.Info("Record updated successfully", "record_id", id)
	c.JSON(http.StatusOK, gin.H{"message": "Record updated", "ID": id})
}
