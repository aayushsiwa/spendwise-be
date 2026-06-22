package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RecalculateBalances(c *gin.Context) {
	if err := h.Service.RefreshBalances(c.Request.Context()); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to recalculate balances", "error", err)
		errors.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Balances recalculated successfully"})
}
