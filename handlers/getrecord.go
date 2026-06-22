package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetRecord(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	rec, err := h.Service.GetRecord(c.Request.Context(), id)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	slog.Info("Record retrieved successfully", "record_id", rec.ID)
	c.JSON(http.StatusOK, rec)
}
