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

func (h *Handler) UpdateBudget(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	var body struct {
		Amount *float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		appErrors.HandleBindingError(c, err, "Invalid request body")
		return
	}

	validationErrs = validator.ValidateUpdateBudgetAmount(body.Amount)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	if err := h.Service.UpdateBudget(c.Request.Context(), id, *body.Amount); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			appErrors.HandleError(c, appErrors.NewNotFound("Budget not found", nil))
			return
		}
		slog.ErrorContext(c.Request.Context(), "Failed to update budget", "error", err)
		appErrors.HandleError(c, appErrors.NewDatabase("Failed to update budget", err))
		return
	}

	slog.InfoContext(c.Request.Context(), "Budget updated", "ID", id)
	c.JSON(http.StatusOK, gin.H{"message": "Budget updated", "ID": id})
}
