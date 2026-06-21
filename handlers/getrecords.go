package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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

	// Build WHERE clause and arguments once (reusable for select + count)
	whereClause, filterArgs := buildWhereClause(queryParams)

	// If groupBy is set, return grouped/aggregated data instead of individual records
	if queryParams.GroupBy != "" {
		h.getGroupedRecords(c, queryParams, whereClause, filterArgs)
		return
	}

	// Base SELECT query
	selectQuery := `
		SELECT r.id, r.date, r.description, COALESCE(c.name, '') as category, r.amount, r.type, r.note, r.balance
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
	` + whereClause + `
		ORDER BY r.date DESC
		LIMIT ? OFFSET ?
	`

	selectArgs := append(append([]any{}, filterArgs...), queryParams.Limit, offset)

	slog.Debug("Executing select query", "query", strings.TrimSpace(selectQuery), "args", selectArgs)

	rows, err := h.DB.Query(selectQuery, selectArgs...)
	if err != nil {
		slog.Error("Failed to execute select query", "error", err)
		appErr := errors.NewDatabase("Failed to retrieve records", err)
		errors.HandleError(c, appErr)
		return
	}
	defer func() { _ = rows.Close() }()

	records := make([]models.Record, 0)

	for rows.Next() {
		var rec models.Record
		if err := rows.Scan(
			&rec.ID,
			&rec.Date,
			&rec.Description,
			&rec.Category,
			&rec.Amount,
			&rec.Type,
			&rec.Note,
			&rec.Balance,
		); err != nil {
			appErr := errors.NewDatabase("Failed to read record data", err)
			errors.HandleError(c, appErr)
			return
		}

		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		appErr := errors.NewDatabase("Error iterating through records", err)
		errors.HandleError(c, appErr)
		return
	}

	// Count query uses same filters but no LIMIT/OFFSET
	countQuery := `
		SELECT COUNT(*)
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
	` + whereClause

	slog.Debug("Executing count query", "query", strings.TrimSpace(countQuery), "args", filterArgs)

	var totalCount int
	if err := h.DB.QueryRow(countQuery, filterArgs...).Scan(&totalCount); err != nil {
		slog.Warn("Failed to get total count", "error", err)
		// Fallback to number of records in current page
		totalCount = len(records)
	}

	// Pagination metadata
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

func (h *Handler) getGroupedRecords(c *gin.Context, q *models.QueryParams, whereClause string, filterArgs []any) {
	var groupExpr, groupAlias string
	switch q.GroupBy {
	case "category":
		groupExpr = "COALESCE(c.name, '')"
		groupAlias = "category"
	case "month":
		groupExpr = "strftime('%Y-%m', r.date)"
		groupAlias = "month"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid groupBy value"})
		return
	}

	query := fmt.Sprintf(`
		SELECT %s AS "%s", SUM(r.amount) AS total, COUNT(*) AS count
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
		%s
		GROUP BY %s
		ORDER BY total DESC
	`, groupExpr, groupAlias, whereClause, groupExpr)

	rows, err := h.DB.Query(query, filterArgs...)
	if err != nil {
		slog.Error("Failed to execute grouped query", "error", err)
		errors.HandleError(c, errors.NewDatabase("Failed to retrieve grouped records", err))
		return
	}
	defer func() { _ = rows.Close() }()

	groups := make([]models.GroupedRecord, 0)
	for rows.Next() {
		var gr models.GroupedRecord
		if err := rows.Scan(&gr.Group, &gr.Total, &gr.Count); err != nil {
			slog.Error("Failed to scan grouped record", "error", err)
			continue
		}
		groups = append(groups, gr)
	}

	c.JSON(http.StatusOK, models.GroupedResponse{Groups: groups})
}
