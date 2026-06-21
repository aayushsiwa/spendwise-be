package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ImportJSON(c *gin.Context) {
	var records []models.Record
	if err := c.ShouldBindJSON(&records); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON array"})
		return
	}

	imported, err := h.Service.ImportJSON(c.Request.Context(), records)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "JSON import failed", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JSON import failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "JSON import completed successfully",
		"recordsImported": imported,
	})
}
