package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
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

	// Base SELECT query
	selectQuery := `
		SELECT r.id, r.date, r.description, c.name as category, r.amount, r.type, r.note, r.balance
		FROM records r
		JOIN categories c ON r.category_id = c.id
	` + whereClause + `
		ORDER BY r.date DESC
		LIMIT ? OFFSET ?
	`

	selectArgs := append(append([]interface{}{}, filterArgs...), queryParams.Limit, offset)

	slog.Debug("Executing select query", "query", strings.TrimSpace(selectQuery), "args", selectArgs)

	rows, err := h.DB.Query(selectQuery, selectArgs...)
	if err != nil {
		slog.Error("Failed to execute select query", "error", err)
		appErr := errors.NewDatabase("Failed to retrieve records", err)
		errors.HandleError(c, appErr)
		return
	}
	defer rows.Close()

	var records []models.Record

	search := strings.ToLower(strings.TrimSpace(queryParams.Search))
	hasSearch := search != ""

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

		// If Description is decrypted before this, keep in-memory search filter:
		if hasSearch {
			if !strings.Contains(strings.ToLower(rec.Description), search) {
				continue
			}
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
		JOIN categories c ON r.category_id = c.id
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
		"filters_applied", whereClause != "" || hasSearch,
		"page", queryParams.Page,
		"limit", queryParams.Limit,
		"total_count", totalCount,
		"total_pages", totalPages,
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

func buildWhereClause(q *models.QueryParams) (string, []interface{}) {
	filters := make([]string, 0, 4)
	args := make([]interface{}, 0, 4)

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

	if len(filters) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(filters, " AND "), args
}
