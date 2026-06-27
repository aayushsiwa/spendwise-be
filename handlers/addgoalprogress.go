package handlers

import (
	"log/slog"
	"net/http"

	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AddGoalProgress(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	var req models.AddProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErrors.HandleBindingError(c, err, "Invalid request body")
		return
	}

	validationErrs = validator.ValidateAddProgress(&req)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	if err := h.Service.AddGoalProgress(c.Request.Context(), id, req.Amount); err != nil {
		appErrors.HandleError(c, err)
		return
	}

	slog.InfoContext(c.Request.Context(), "Goal progress added", "ID", id, "amount", req.Amount)
	c.JSON(http.StatusOK, gin.H{"message": "Progress added", "ID": id})
}
