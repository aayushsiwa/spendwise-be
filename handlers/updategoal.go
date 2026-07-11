package handlers

import (
	"log/slog"
	"net/http"

	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateGoal(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	var req models.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErrors.HandleBindingError(c, err, "Invalid request body")
		return
	}

	validationErrs = validator.ValidateUpdateGoal(&req)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	if err := h.Service.UpdateGoal(c.Request.Context(), id, &req); err != nil {
		appErrors.HandleError(c, err)
		return
	}

	slog.InfoContext(c.Request.Context(), "Goal updated", "ID", id)
	c.JSON(http.StatusOK, gin.H{"message": "Goal updated", "ID": id})
}
