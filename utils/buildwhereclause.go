package utils

import (
	"aayushsiwa/expense-tracker/models"
	"strings"
)

func BuildWhereClause(q *models.RecordsQueryParams) (string, []interface{}) {
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
