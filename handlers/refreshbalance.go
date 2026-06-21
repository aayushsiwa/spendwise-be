package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RecalculateBalances(c *gin.Context) {
	if err := h.Service.RefreshBalances(c.Request.Context()); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to recalculate balances", "error", err)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(200, gin.H{"status": "Balances recalculated successfully"})
}
