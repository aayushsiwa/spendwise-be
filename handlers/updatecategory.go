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
	var count int64
	err := h.DB.Model(&models.Category{}).Where(`"ID" = ?`, id).Count(&count).Error
	if err != nil {
		appErr := errors.NewDatabase("Failed to check category existence", err)
		errors.HandleError(c, appErr)
		return
	}

	if count == 0 {
		appErr := errors.NewNotFound("Category not found", nil).WithDetails(map[string]any{
			"categoryID": id,
		})
		errors.HandleError(c, appErr)
		return
	}

	err = h.DB.Model(&models.Category{}).Where(`"ID" = ?`, id).Updates(models.Category{
		Name:  cat.Name,
		Icon:  cat.Icon,
		Color: cat.Color,
	}).Error
	if err != nil {
		appErr := errors.NewDatabase("Failed to update category", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Category updated successfully", "categoryID", id, "name", cat.Name)
	c.JSON(http.StatusOK, gin.H{"ID": id, "name": cat.Name, "icon": cat.Icon, "color": cat.Color})
}
