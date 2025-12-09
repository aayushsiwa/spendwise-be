package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	var cat models.Category
	if err := c.BindJSON(&cat); err != nil {
		appErr := errors.NewInvalidInput("Invalid JSON body", err)
		errors.HandleError(c, appErr)
		return
	}

	// Validate category data
	validationErrs = validator.ValidateCategory(&cat)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Check if category exists
	var exists int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE id = ?", id).Scan(&exists)
	if err != nil {
		appErr := errors.NewDatabase("Failed to check category existence", err)
		errors.HandleError(c, appErr)
		return
	}

	if exists == 0 {
		appErr := errors.NewNotFound("Category not found", nil).WithDetails(map[string]interface{}{
			"category_id": id,
		})
		errors.HandleError(c, appErr)
		return
	}

	_, err = h.DB.Exec("UPDATE categories SET name = ?, icon = ?, color = ? WHERE id = ?",
		cat.Name, cat.Icon, cat.Color, id)
	if err != nil {
		appErr := errors.NewDatabase("Failed to update category", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Category updated successfully", "category_id", id, "name", cat.Name)
	c.JSON(http.StatusOK, gin.H{"id": id, "name": cat.Name, "icon": cat.Icon, "color": cat.Color})
}
