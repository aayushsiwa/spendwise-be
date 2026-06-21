package handlers

import (
	"log/slog"
	"net/http"
	"strings"

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
		appErr := errors.NewInvalidInput("Invalid query parameters", err)
		errors.HandleError(c, appErr)
		return
	}

	offset := (queryParams.Page - 1) * queryParams.Limit
	whereClause, filterArgs := buildWhereClause(queryParams)

	if queryParams.GroupBy != "" {
		groups, err := h.Service.GetGroupedRecords(c.Request.Context(), queryParams.GroupBy, whereClause, filterArgs)
		if err != nil {
			errors.HandleError(c, err)
			return
		}
		c.JSON(http.StatusOK, models.GroupedResponse{Groups: groups})
		return
	}

	records, totalCount, err := h.Service.GetRecords(c.Request.Context(), whereClause, filterArgs, queryParams.Limit, offset)
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

	slog.Info("Records retrieved successfully",
		"count", len(records),
		"filtersApplied", whereClause != "",
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

func buildWhereClause(q *models.QueryParams) (string, []any) {
	filters := make([]string, 0, 5)
	args := make([]any, 0, 5)

	if q.Type != "" {
		filters = append(filters, "r.type = ?")
		args = append(args, q.Type)
	}
	if q.Category != "" {
		filters = append(filters, "c.name = ?")
		args = append(args, q.Category)
	}
	if q.From != "" {
		filters = append(filters, "r.date >= ?")
		args = append(args, q.From)
	}
	if q.To != "" {
		filters = append(filters, "r.date <= ?")
		args = append(args, q.To)
	}
	if q.MinAmount != 0 {
		filters = append(filters, "r.amount >= ?")
		args = append(args, q.MinAmount)
	}
	if q.MaxAmount != 0 {
		filters = append(filters, "r.amount <= ?")
		args = append(args, q.MaxAmount)
	}
	if q.Search != "" {
		filters = append(filters, "LOWER(r.description) LIKE ?")
		args = append(args, "%"+strings.ToLower(q.Search)+"%")
	}

	if len(filters) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(filters, " AND "), args
}
