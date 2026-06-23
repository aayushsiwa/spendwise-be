package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	err := h.Service.DeleteCategory(c.Request.Context(), id)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	slog.InfoContext(c.Request.Context(), "Category deleted successfully", "categoryID", id)
	c.JSON(http.StatusNoContent, nil)
}
