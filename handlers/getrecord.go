package handlers

import (
	appErrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) GetRecord(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		appErrors.HandleValidationErrors(c, validationErrs)
		return
	}

	var rec models.Record
	err := h.DB.Table("records r").
		Select(`r.id, r.date, r.description, COALESCE(c.name, '') as category, r.amount, r.type, r.note, r.balance`).
		Joins(`LEFT JOIN categories c ON r."categoryID" = c.ID`).
		Where("r.id = ?", id).
		Scan(&rec).Error

	if err != nil {
		appErr := appErrors.NewDatabase("Failed to read record", err)
		appErrors.HandleError(c, appErr)
		return
	}

	if rec.ID == "" {
		appErr := appErrors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), gorm.ErrRecordNotFound)
		appErrors.HandleError(c, appErr)
		return
	}

	slog.Info("Record retrieved successfully", "record_id", rec.ID)
	c.JSON(http.StatusOK, rec)
}
