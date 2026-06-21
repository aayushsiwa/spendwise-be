package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateRecord(c *gin.Context) {
	var rec models.Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		appErr := errors.NewInvalidInput("Invalid JSON body", err)
		errors.HandleError(c, appErr)
		return
	}

	validator := validation.NewValidator()
	validationErrs := validator.ValidateRecord(&rec)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	customId, err := h.GenerateCustomID(rec.Date)
	if err != nil {
		appErr := errors.NewInternal("Failed to generate record ID", err)
		errors.HandleError(c, appErr)
		return
	}

	rec.ID = customId

	if err := h.Service.CreateRecord(c.Request.Context(), &rec); err != nil {
		errors.HandleError(c, err)
		return
	}

	slog.Info("Record created successfully", "record_id", rec.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Record with id " + rec.ID + " created successfully",
		"ID":      rec.ID,
	})
}
