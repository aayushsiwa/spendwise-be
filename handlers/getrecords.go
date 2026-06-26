package handlers

import (
	"log/slog"
	"net/http"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetRecords(c *gin.Context) {
	queryParams := &models.QueryParams{
		PaginationFilterParams: models.PaginationFilterParams{
			Limit: 10,
			Page:  1,
		},
	}

	if err := c.ShouldBindQuery(queryParams); err != nil {
		errors.HandleBindingError(c, err, "Invalid query parameters")
		return
	}

	if queryParams.GroupBy != "" {
		groups, err := h.Service.GetGroupedRecords(c.Request.Context(), queryParams)
		if err != nil {
			errors.HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, models.GroupedResponse{Groups: groups})
		return
	}

	records, totalCount, err := h.Service.GetRecords(c.Request.Context(), queryParams)
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	totalPages := 0
	if queryParams.Limit > 0 {
		totalPages = (totalCount + queryParams.Limit - 1) / queryParams.Limit
	}
	hasNext := queryParams.Page < totalPages
	hasPrev := queryParams.Page > 1

	slog.InfoContext(c.Request.Context(), "Records retrieved successfully",
		"count", len(records),
		"page", queryParams.Page,
		"limit", queryParams.Limit,
		"totalCount", totalCount,
		"totalPages", totalPages,
	)

	res := models.RecordsResponse{
		Records: records,
		PaginationMetadata: models.PaginationMetadata{
			Page:       queryParams.Page,
			Limit:      queryParams.Limit,
			TotalCount: totalCount,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}

	c.JSON(http.StatusOK, res)
}
