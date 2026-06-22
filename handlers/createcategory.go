package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"

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

	validator := validation.NewValidator()
	var allValidationErrs errors.ValidationErrors

	for i, cat := range categories {
		validationErrs := validator.ValidateCategory(&cat)
		for _, err := range validationErrs {
			err.Field = fmt.Sprintf("categories[%d].%s", i, err.Field)
			allValidationErrs = append(allValidationErrs, err)
		}
	}

	if len(allValidationErrs) > 0 {
		errors.HandleValidationErrors(c, allValidationErrs)
		return
	}

	inserted, err := h.Service.CreateCategories(c.Request.Context(), categories)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	insertedRes := make([]gin.H, 0, len(inserted))
	for _, cat := range inserted {
		insertedRes = append(insertedRes, gin.H{
			"ID":    cat.ID,
			"name":  cat.Name,
			"icon":  cat.Icon,
			"color": cat.Color,
		})
	}

<<<<<<< HEAD
	if err := tx.Commit().Error; err != nil {
		appErr := errors.NewDatabase("Failed to commit transaction", err)
		errors.HandleError(c, appErr)
		return
	}

=======
>>>>>>> 0617a2afde94cf5b86ce3dd3494faae90d7b64cd
	slog.Info("Categories created successfully", "count", len(inserted))
	c.JSON(http.StatusCreated, insertedRes)
}
