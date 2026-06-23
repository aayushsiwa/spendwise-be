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
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

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
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	progress, err := h.Service.GetBudgetProgress(c.Request.Context(), month, year)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to get budget progress", "error", err)
		appErrors.HandleError(c, appErrors.NewDatabase("Failed to retrieve budget progress", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"progress": progress})
}
