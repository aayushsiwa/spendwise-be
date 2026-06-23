package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	var cat models.Category
	if err := c.BindJSON(&cat); err != nil {
		errors.HandleBindingError(c, err, "Invalid JSON body")
		return
	}

	validationErrs = validator.ValidateCategory(&cat)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	if err := h.Service.UpdateCategory(c.Request.Context(), id, &cat); err != nil {
		errors.HandleError(c, err)
		return
	}

	slog.InfoContext(c.Request.Context(), "Category updated successfully", "categoryID", id, "name", cat.Name)
	c.JSON(http.StatusOK, gin.H{"ID": id, "name": cat.Name, "icon": cat.Icon, "color": cat.Color})
}
