package handlers

import (
	"log/slog"
	"net/http"

	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateGoal(c *gin.Context) {
	var goal models.Goal
	if err := c.ShouldBindJSON(&goal); err != nil {
		appErrors.HandleBindingError(c, err, "Invalid goal data")
		return
	}

	validator := validation.NewValidator()
	validationErrs := validator.ValidateGoal(&goal)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	if err := h.Service.CreateGoal(c.Request.Context(), &goal); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to create goal", "error", err)
		appErrors.HandleError(c, err)
		return
	}

	slog.InfoContext(c.Request.Context(), "Goal created", "ID", goal.ID)
	c.JSON(http.StatusCreated, gin.H{"message": "Goal created", "ID": goal.ID})
}
