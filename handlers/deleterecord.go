package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func (h *Handler) DeleteRecord(c *gin.Context) {
	idStr := c.Param("id")

	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	_, err := h.Service.DeleteRecord(c.Request.Context(), id)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	slog.InfoContext(c.Request.Context(), "Record deleted successfully", "record_id", id)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Record with id %s deleted successfully", id),
	})
}
