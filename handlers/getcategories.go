package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCategories(c *gin.Context) {
	rows, err := h.DB.Query("SELECT id, name, icon, color FROM categories ORDER BY name ASC")
	if err != nil {
		appErr := errors.NewDatabase("Failed to fetch categories", err)
		errors.HandleError(c, appErr)
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Icon, &cat.Color); err != nil {
			slog.Warn("Failed to scan category row", "error", err)
			continue // Skip invalid rows but continue processing
		}
		categories = append(categories, cat)
	}

	if err = rows.Err(); err != nil {
		appErr := errors.NewDatabase("Error iterating through categories", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Categories retrieved successfully", "count", len(categories))
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
