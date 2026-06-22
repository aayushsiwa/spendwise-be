package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCategories(c *gin.Context) {
	categories, err := h.Service.GetCategories(c.Request.Context())
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	slog.Info("Categories retrieved successfully", "count", len(categories))
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
