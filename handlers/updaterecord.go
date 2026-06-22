package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) PatchRecord(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	var req models.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErrors.HandleError(c, appErrors.NewInvalidInput("Invalid JSON body", err))
		return
	}

	validationErrs = validator.ValidatePatchRecord(&req)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	if err := h.Service.PatchRecord(c.Request.Context(), id, &req); err != nil {
		errors.HandleError(c, err)
		return
	}

	slog.Info("Record updated successfully", "record_id", id)
	c.JSON(http.StatusOK, gin.H{"message": "Record updated", "ID": id})
}
