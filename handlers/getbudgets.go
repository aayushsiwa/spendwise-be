package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	appErrors "aayushsiwa/expense-tracker/errors"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetBudgets(c *gin.Context) {
	now := time.Now()
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(now.Month())))
	yearStr := c.DefaultQuery("year", strconv.Itoa(now.Year()))

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		appErrors.HandleValidationErrors(c, appErrors.ValidationErrors{
			appErrors.NewValidationError("month", "must be a number", monthStr),
		})
		return
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		appErrors.HandleValidationErrors(c, appErrors.ValidationErrors{
			appErrors.NewValidationError("year", "must be a number", yearStr),
		})
		return
	}

	if month < 1 || month > 12 {
		appErrors.HandleValidationErrors(c, appErrors.ValidationErrors{
			appErrors.NewValidationError("month", "must be between 1 and 12", month),
		})
		return
	}

	if year < 1 {
		appErrors.HandleValidationErrors(c, appErrors.ValidationErrors{
			appErrors.NewValidationError("year", "must be a positive number", year),
		})
		return
	}

	budgets, err := h.Service.GetBudgets(c.Request.Context(), month, year)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to get budgets", "error", err)
		appErrors.HandleError(c, appErrors.NewDatabase("Failed to retrieve budgets", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"budgets": budgets})
}

func (h *Handler) GetBudgetProgress(c *gin.Context) {
	now := time.Now()
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(now.Month())))
	yearStr := c.DefaultQuery("year", strconv.Itoa(now.Year()))

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		appErrors.HandleValidationErrors(c, appErrors.ValidationErrors{
			appErrors.NewValidationError("month", "must be a number", monthStr),
		})
		return
	}
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		appErrors.HandleValidationErrors(c, appErrors.ValidationErrors{
			appErrors.NewValidationError("year", "must be a number", yearStr),
		})
		return
	}

	if month < 1 || month > 12 {
		appErrors.HandleValidationErrors(c, appErrors.ValidationErrors{
			appErrors.NewValidationError("month", "must be between 1 and 12", month),
		})
		return
	}

	if year < 1 {
		appErrors.HandleValidationErrors(c, appErrors.ValidationErrors{
			appErrors.NewValidationError("year", "must be a positive number", year),
		})
		return
	}

	progress, err := h.Service.GetBudgetProgress(c.Request.Context(), month, year)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to get budget progress", "error", err)
		appErrors.HandleError(c, appErrors.NewDatabase("Failed to retrieve budget progress", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"progress": progress})
}
