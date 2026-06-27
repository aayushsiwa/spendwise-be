package handlers

import (
	"log/slog"
	"net/http"

	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteGoal(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	if err := h.Service.DeleteGoal(c.Request.Context(), id); err != nil {
		appErrors.HandleError(c, err)
		return
	}

	slog.InfoContext(c.Request.Context(), "Goal deleted", "ID", id)
	c.JSON(http.StatusOK, gin.H{"message": "Goal deleted", "ID": id})
}
