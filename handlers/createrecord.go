package handlers

import (
	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateRecord(c *gin.Context) {
	var rec models.Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		appErr := appErrors.NewInvalidInput("Invalid JSON body", err)
		appErrors.HandleError(c, appErr)
		return
	}

	validator := validation.NewValidator()
	validationErrs := validator.ValidateRecord(&rec)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	customId, err := h.GenerateCustomID(rec.Date)
	if err != nil {
		appErr := appErrors.NewInternal("Failed to generate record ID", err)
		appErrors.HandleError(c, appErr)
		return
	}

	rec.ID = customId

	if err := h.Service.CreateRecord(c.Request.Context(), &rec); err != nil {
		appErrors.HandleError(c, err)
		return
	}

	slog.InfoContext(c.Request.Context(), "Record created successfully", "record_id", rec.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Record with id " + rec.ID + " created successfully",
		"ID":      rec.ID,
	})
}
