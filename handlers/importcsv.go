package handlers

import (
	"errors"
	"net/http"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/services"

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

	file, err := openFileFunc(fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer func() { _ = file.Close() }()

	imported, skipped, err := h.Service.ImportCSV(c.Request.Context(), file)
	if err != nil {
		if errors.Is(err, services.ErrImportValidation) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		apperrors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "CSV import completed successfully",
		"recordsImported": imported,
		"skippedCount":    skipped,
	})
}
