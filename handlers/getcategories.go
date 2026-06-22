package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCategories(c *gin.Context) {
	var categories []models.Category
	err := h.DB.Order("name ASC").Find(&categories).Error
	if err != nil {
		appErr := errors.NewDatabase("Failed to fetch categories", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Categories retrieved successfully", "count", len(categories))
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
