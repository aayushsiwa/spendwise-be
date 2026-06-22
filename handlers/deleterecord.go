package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
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

	result := h.DB.Where("id = ?", id).Delete(&models.Record{})
	if result.Error != nil {
		appErr := errors.NewDatabase("Failed to delete record", result.Error)
		errors.HandleError(c, appErr)
		return
	}

	if result.RowsAffected == 0 {
		appErr := errors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), nil)
		errors.HandleError(c, appErr)
		return
	}

	// Update summary
	if err := h.UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record deletion", "record_id", id, "error", err)
	}

	slog.Info("Record deleted successfully", "record_id", id)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Record with id %s deleted successfully", id),
	})
}
