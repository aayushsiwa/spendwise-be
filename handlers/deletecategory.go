package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
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

	// Check if category is being used by any records
	var recordCount int64
	err = h.DB.Model(&models.Record{}).Where(`"categoryID" = ?`, id).Count(&recordCount).Error
	if err != nil {
		appErr := errors.NewDatabase("Failed to check category usage", err)
		errors.HandleError(c, appErr)
		return
	}

	if recordCount > 0 {
		appErr := errors.NewConflict("Cannot delete category that has associated records", nil).WithDetails(map[string]any{
			"categoryID":  id,
			"recordCount": recordCount,
		})
		errors.HandleError(c, appErr)
		return
	}

	err = h.DB.Where(`"ID" = ?`, id).Delete(&models.Category{}).Error
	if err != nil {
		appErr := errors.NewDatabase("Failed to delete category", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Category deleted successfully", "categoryID", id)
	c.JSON(http.StatusNoContent, nil)
}
