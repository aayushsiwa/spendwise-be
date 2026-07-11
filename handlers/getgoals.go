package handlers

import (
	"log/slog"
	"net/http"

	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetGoals(c *gin.Context) {
	goals, err := h.Service.GetGoals(c.Request.Context())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to get goals", "error", err)
		appErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"goals": goals})
}

func (h *Handler) GetGoal(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	goal, err := h.Service.GetGoal(c.Request.Context(), id)
	if err != nil {
		appErrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"goal": goal})
}
