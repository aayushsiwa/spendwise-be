package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v4"
)

func (h *Handler) CreateCategories(c *gin.Context) {
	var categories []models.Category

	if err := c.BindJSON(&categories); err != nil {
		appErr := errors.NewInvalidInput("Invalid JSON body", err)
		errors.HandleError(c, appErr)
		return
	}

	if len(categories) == 0 {
		appErr := errors.NewInvalidInput("No categories provided", nil)
		errors.HandleError(c, appErr)
		return
	}

	// Validate all categories before processing
	validator := validation.NewValidator()
	var allValidationErrs errors.ValidationErrors

	for i, cat := range categories {
		validationErrs := validator.ValidateCategory(&cat)
		for _, err := range validationErrs {
			// Add index to field name for better error reporting
			err.Field = fmt.Sprintf("categories[%d].%s", i, err.Field)
			allValidationErrs = append(allValidationErrs, err)
		}
	}

	if len(allValidationErrs) > 0 {
		errors.HandleValidationErrors(c, allValidationErrs)
		return
	}

	tx := h.DB.Begin()
	if tx.Error != nil {
		appErr := errors.NewDatabase("Failed to begin transaction", tx.Error)
		errors.HandleError(c, appErr)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var inserted []gin.H
	for _, cat := range categories {
		if cat.Name == "" {
			continue
		}
		catID := shortuuid.New()
		lowerName := strings.ToLower(cat.Name)

		newCat := models.Category{
			ID:    catID,
			Name:  lowerName,
			Icon:  cat.Icon,
			Color: cat.Color,
		}

		if err := tx.Create(&newCat).Error; err != nil {
			tx.Rollback()
			appErr := errors.NewDatabase("Failed to insert category", err).WithDetails(map[string]any{
				"categoryName": cat.Name,
			})
			errors.HandleError(c, appErr)
			return
		}

		inserted = append(inserted, gin.H{
			"ID":    catID,
			"name":  lowerName,
			"icon":  cat.Icon,
			"color": cat.Color,
		})
	}

	if err := tx.Commit().Error; err != nil {
		appErr := errors.NewDatabase("Failed to commit transaction", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Categories created successfully", "count", len(inserted))
	c.JSON(http.StatusCreated, inserted)
}
