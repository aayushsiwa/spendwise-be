package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCategoryRecords(c *gin.Context) {
	pathParts := strings.Split(c.Request.URL.Path, "/")
	if len(pathParts) < 2 {
		appErr := errors.NewInvalidInput("Invalid URL path", nil)
		errors.HandleError(c, appErr)
		return
	}
	categoryId := pathParts[len(pathParts)-1]

	// Validate category ID
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(categoryId)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	var categoryName string
	err := h.DB.QueryRow("SELECT name FROM categories WHERE id = ?", id).Scan(&categoryName)
	if err != nil {
		appErr := errors.NewNotFound("Category not found", err).WithDetails(map[string]interface{}{
			"category_id": id,
		})
		errors.HandleError(c, appErr)
		return
	}

	rows, err := h.DB.Query(`
		SELECT r.id, r.date, r.description, r.category_id, c.name, r.amount, r.type, r.note
		FROM records r
		JOIN categories c ON r.category_id = c.id
		WHERE c.name = ?
		ORDER BY r.date DESC
	`, categoryName)
	if err != nil {
		appErr := errors.NewDatabase("Failed to fetch records", err)
		errors.HandleError(c, appErr)
		return
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var r models.Record
		err := rows.Scan(&r.ID, &r.Date, &r.Description, &categoryId, &r.Category, &r.Amount, &r.Type, &r.Note)
		if err != nil {
			slog.Warn("Failed to scan record row", "error", err)
			continue // Skip invalid rows but continue processing
		}
		records = append(records, r)
	}

	if err = rows.Err(); err != nil {
		appErr := errors.NewDatabase("Error iterating through records", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Category records retrieved successfully", "category", categoryName, "count", len(records))
	c.JSON(http.StatusOK, records)
}
