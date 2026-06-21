package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ImportCSV(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file not provided"})
		return
	}

	if fileHeader.Size > 10<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer func() { _ = file.Close() }()

	imported, skipped, err := h.Service.ImportCSV(c.Request.Context(), file)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "CSV import failed", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "CSV import failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "CSV import completed successfully",
		"recordsImported": imported,
		"skippedCount":    skipped,
	})
}
