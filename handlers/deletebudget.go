package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) DeleteBudget(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	if err := h.Service.DeleteBudget(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErrors.HandleError(c, appErrors.NewNotFound("Budget not found", nil))
			return
		}
		slog.ErrorContext(c.Request.Context(), "Failed to delete budget", "error", err)
		appErrors.HandleError(c, appErrors.NewDatabase("Failed to delete budget", err))
		return
	}

	slog.InfoContext(c.Request.Context(), "Budget deleted", "ID", id)
	c.JSON(http.StatusOK, gin.H{"message": "Budget deleted", "ID": id})
}
