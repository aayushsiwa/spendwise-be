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

	tx, err := h.DB.Begin()
	if err != nil {
		appErr := errors.NewDatabase("Failed to begin transaction", err)
		errors.HandleError(c, appErr)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO categories (name, icon, color) VALUES (?, ?, ?)")
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			slog.Error("Failed to rollback transaction", "error", rbErr)
		}
		appErr := errors.NewDatabase("Failed to prepare statement", err)
		errors.HandleError(c, appErr)
		return
	}
	defer stmt.Close()

	var inserted []gin.H
	for _, cat := range categories {
		if cat.Name == "" {
			continue
		}
		lowerName := strings.ToLower(cat.Name)
		result, err := stmt.Exec(lowerName, cat.Icon, cat.Color)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.Error("Failed to rollback transaction", "error", rbErr)
			}
			appErr := errors.NewDatabase("Failed to insert category", err).WithDetails(map[string]interface{}{
				"category_name": cat.Name,
			})
			errors.HandleError(c, appErr)
			return
		}
		id, _ := result.LastInsertId()
		inserted = append(inserted, gin.H{
			"id":    id,
			"name":  lowerName,
			"icon":  cat.Icon,
			"color": cat.Color,
		})
	}

	if err := tx.Commit(); err != nil {
		appErr := errors.NewDatabase("Failed to commit transaction", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Categories created successfully", "count", len(inserted))
	c.JSON(http.StatusCreated, inserted)
}
