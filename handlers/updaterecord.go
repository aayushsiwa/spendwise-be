package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PatchRecord(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	var req models.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, errors.NewInvalidInput("Invalid JSON body", err))
		return
	}

	validationErrs = validator.ValidatePatchRecord(&req)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	if err := h.Service.PatchRecord(c.Request.Context(), id, &req); err != nil {
		errors.HandleError(c, err)
		return
	}

	slog.Info("Record updated successfully", "record_id", id)
	c.JSON(http.StatusOK, gin.H{"message": "Record updated", "ID": id})
}
