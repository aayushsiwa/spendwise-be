package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteRecord(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	res, err := h.DB.Exec(`DELETE FROM records WHERE id = ?`, id)
	if err != nil {
		appErr := errors.NewDatabase("Failed to delete record", err)
		errors.HandleError(c, appErr)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		appErr := errors.NewDatabase("Failed to get affected rows", err)
		errors.HandleError(c, appErr)
		return
	}

	if rowsAffected == 0 {
		appErr := errors.NewNotFound(fmt.Sprintf("Record with ID %d not found", id), nil)
		errors.HandleError(c, appErr)
		return
	}

	// Update summary
	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record deletion", "record_id", id, "error", err)
	}

	slog.Info("Record deleted successfully", "record_id", id)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Record with id %d deleted successfully", id),
	})
}
