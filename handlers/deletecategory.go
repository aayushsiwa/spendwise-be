package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/validation"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
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

	// Check if category is being used by any records
	var recordCount int
	err = h.DB.QueryRow("SELECT COUNT(*) FROM records WHERE category_id = ?", id).Scan(&recordCount)
	if err != nil {
		appErr := errors.NewDatabase("Failed to check category usage", err)
		errors.HandleError(c, appErr)
		return
	}

	if recordCount > 0 {
		appErr := errors.NewConflict("Cannot delete category that has associated records", nil).WithDetails(map[string]interface{}{
			"category_id":  id,
			"record_count": recordCount,
		})
		errors.HandleError(c, appErr)
		return
	}

	_, err = h.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		appErr := errors.NewDatabase("Failed to delete category", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Category deleted successfully", "category_id", id)
	c.JSON(http.StatusNoContent, nil)
}
