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

func (h *Handler) PatchRecord(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Check if record exists
	var count int64
	err := h.DB.Model(&models.Record{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		appErrors.HandleError(c, appErrors.NewDatabase("Failed to check record existence", err))
		return
	}
	if count == 0 {
		appErrors.HandleError(c, appErrors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), nil))
		return
	}

	var req models.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErrors.HandleError(c, appErrors.NewInvalidInput("Invalid JSON body", err))
		return
	}

	validationErrs = validator.ValidatePatchRecord(&req)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
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
		var category models.Category
		err := h.DB.Select(`"ID"`).Where("name = ?", *req.Category).First(&category).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				appErr := appErrors.NewInvalidInput("Category not found", err).WithDetails(map[string]any{
					"category": *req.Category,
				})
				appErrors.HandleError(c, appErr)
			} else {
				appErrors.HandleError(c, appErrors.NewDatabase("Failed to find category", err))
			}
			return
		}
		setClauses = append(setClauses, `"categoryID" = ?`)
		args = append(args, category.ID)
	}

	if len(setClauses) == 0 {
		slog.Info("No fields to update, returning success (idempotent)", "record_id", id)
		c.JSON(http.StatusOK, gin.H{"message": "No fields to update", "ID": id})
		return
	}

	query := fmt.Sprintf("UPDATE records SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)

	err = h.DB.Exec(query, args...).Error
	if err != nil {
		appErrors.HandleError(c, appErrors.NewDatabase("Failed to update record", err))
		return
	}

	// Recalculate all balances
	tx := h.DB.Begin()
	if tx.Error != nil {
		slog.Warn("Failed to start transaction for balance recalculation", "error", tx.Error)
	} else {
		if err := h.recalculateBalances(ctx, tx); err != nil {
			slog.Warn("Failed to recalculate balances after record update", "record_id", id, "error", err)
			tx.Rollback()
		} else {
			if cErr := tx.Commit().Error; cErr != nil {
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
