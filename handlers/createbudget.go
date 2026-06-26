package handlers

import (
	"log/slog"
	"net/http"

	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v4"
)

func (h *Handler) CreateBudget(c *gin.Context) {
	var budget models.Budget
	if err := c.ShouldBindJSON(&budget); err != nil {
		appErrors.HandleBindingError(c, err, "Invalid budget data")
		return
	}

	validator := validation.NewValidator()
	validationErrs := validator.ValidateBudget(&budget)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	budget.ID = shortuuid.New()

	if err := h.Service.CreateBudget(c.Request.Context(), &budget); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to create budget", "error", err)
		appErrors.HandleError(c, appErrors.NewDatabase("Failed to create budget", err))
		return
	}

	slog.InfoContext(c.Request.Context(), "Budget created", "ID", budget.ID)

	c.JSON(http.StatusCreated, gin.H{"message": "Budget created", "ID": budget.ID})
}
