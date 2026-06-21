package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
)

func (s *RecordService) CreateRecord(ctx context.Context, rec *models.Record) error {
	var categoryID string
	err := s.db.QueryRowContext(ctx, `SELECT "ID" FROM categories WHERE name = ?`, strings.ToLower(rec.Category)).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.NewInvalidInput("Category not found", err).WithDetails(map[string]any{
				"category": rec.Category,
			})
		}
		return errors.NewDatabase("Failed to find category", err)
	}

	var currentBalance float64
	err = s.db.QueryRowContext(ctx, "SELECT COALESCE(balance, 0) FROM records ORDER BY date DESC, id DESC LIMIT 1").Scan(&currentBalance)
	if err != nil {
		currentBalance = 0
	}

	switch rec.Type {
	case "income":
		currentBalance += rec.Amount
	case "expense":
		currentBalance -= rec.Amount
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO records (id, date, description, "categoryID", amount, type, note, balance)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.Date, rec.Description, categoryID, rec.Amount, rec.Type, rec.Note, currentBalance)
	if err != nil {
		return errors.NewDatabase("Failed to insert record", err)
	}

	if err := s.UpdateSummary(ctx); err != nil {
		slog.ErrorContext(ctx, "Failed to update summary after record creation", "record_id", rec.ID, "error", err)
	}

	return nil
}

func (s *RecordService) GetRecord(ctx context.Context, id string) (*models.Record, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT r.id, r.date, r.description, COALESCE(c.name, '') as category, r.amount, r.type, r.note, r.balance
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
		WHERE r.id = ?
	`, id)

	var rec models.Record
	if err := row.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Note, &rec.Balance); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), err)
		}
		return nil, errors.NewDatabase("Failed to read record", err)
	}
	return &rec, nil
}

func (s *RecordService) GetRecords(ctx context.Context, whereClause string, filterArgs []any, limit, offset int) ([]models.Record, int, error) {
	selectQuery := `
		SELECT r.id, r.date, r.description, COALESCE(c.name, '') as category, r.amount, r.type, r.note, r.balance
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
	` + whereClause + `
		ORDER BY r.date DESC
		LIMIT ? OFFSET ?
	`

	selectArgs := append(append([]any{}, filterArgs...), limit, offset)

	rows, err := s.db.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, errors.NewDatabase("Failed to retrieve records", err)
	}
	defer func() { _ = rows.Close() }()

	records := make([]models.Record, 0)
	for rows.Next() {
		var rec models.Record
		if err := rows.Scan(
			&rec.ID, &rec.Date, &rec.Description, &rec.Category,
			&rec.Amount, &rec.Type, &rec.Note, &rec.Balance,
		); err != nil {
			return nil, 0, errors.NewDatabase("Failed to read record data", err)
		}
		records = append(records, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, errors.NewDatabase("Error iterating through records", err)
	}

	countQuery := `
		SELECT COUNT(*)
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
	` + whereClause

	var totalCount int
	if err := s.db.QueryRowContext(ctx, countQuery, filterArgs...).Scan(&totalCount); err != nil {
		slog.WarnContext(ctx, "Failed to get total count", "error", err)
		totalCount = len(records)
	}

	return records, totalCount, nil
}

func (s *RecordService) GetGroupedRecords(ctx context.Context, groupBy, whereClause string, filterArgs []any) ([]models.GroupedRecord, error) {
	var groupExpr, groupAlias string
	switch groupBy {
	case "category":
		groupExpr = "COALESCE(c.name, '')"
		groupAlias = "category"
	case "month":
		groupExpr = "strftime('%Y-%m', r.date)"
		groupAlias = "month"
	default:
		return nil, errors.NewInvalidInput("Invalid groupBy value", nil)
	}

	query := fmt.Sprintf(`
		SELECT %s AS "%s", SUM(r.amount) AS total, COUNT(*) AS count
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
		%s
		GROUP BY %s
		ORDER BY total DESC
	`, groupExpr, groupAlias, whereClause, groupExpr)

	rows, err := s.db.QueryContext(ctx, query, filterArgs...)
	if err != nil {
		return nil, errors.NewDatabase("Failed to retrieve grouped records", err)
	}
	defer func() { _ = rows.Close() }()

	groups := make([]models.GroupedRecord, 0)
	for rows.Next() {
		var gr models.GroupedRecord
		if err := rows.Scan(&gr.Group, &gr.Total, &gr.Count); err != nil {
			slog.ErrorContext(ctx, "Failed to scan grouped record", "error", err)
			continue
		}
		groups = append(groups, gr)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabase("Error iterating grouped records", err)
	}

	return groups, nil
}

func (s *RecordService) PatchRecord(ctx context.Context, id string, req *models.UpdateRecordRequest) error {
	var exists int
	err := s.db.QueryRow("SELECT COUNT(*) FROM records WHERE id = ?", id).Scan(&exists)
	if err != nil {
		return errors.NewDatabase("Failed to check record existence", err)
	}
	if exists == 0 {
		return errors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), nil)
	}

	var setClauses []string
	var args []any

	if req.Date != nil {
		setClauses = append(setClauses, "date = ?")
		args = append(args, *req.Date)
	}
	if req.Description != nil {
		setClauses = append(setClauses, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Amount != nil {
		setClauses = append(setClauses, "amount = ?")
		args = append(args, *req.Amount)
	}
	if req.Type != nil {
		setClauses = append(setClauses, "type = ?")
		args = append(args, *req.Type)
	}
	if req.Note != nil {
		setClauses = append(setClauses, "note = ?")
		args = append(args, *req.Note)
	}
	if req.Category != nil {
		var categoryID string
		err := s.db.QueryRow(`SELECT "ID" FROM categories WHERE name = ?`, *req.Category).Scan(&categoryID)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.NewInvalidInput("Category not found", err).WithDetails(map[string]any{
					"category": *req.Category,
				})
			}
			return errors.NewDatabase("Failed to find category", err)
		}
		setClauses = append(setClauses, `"categoryID" = ?`)
		args = append(args, categoryID)
	}

	if len(setClauses) == 0 {
		return nil
	}

	query := fmt.Sprintf("UPDATE records SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.NewDatabase("Failed to update record", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		slog.WarnContext(ctx, "Failed to start transaction for balance recalculation", "error", err)
	} else {
		if err := recalculateBalances(ctx, tx); err != nil {
			slog.WarnContext(ctx, "Failed to recalculate balances after record update", "record_id", id, "error", err)
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.ErrorContext(ctx, "Failed to rollback transaction", "error", rbErr)
			}
		} else {
			if cErr := tx.Commit(); cErr != nil {
				slog.ErrorContext(ctx, "Failed to commit transaction", "error", cErr)
			}
		}
	}

	if err := s.UpdateSummary(ctx); err != nil {
		slog.ErrorContext(ctx, "Failed to update summary after record update", "record_id", id, "error", err)
	}

	return nil
}

func (s *RecordService) DeleteRecord(ctx context.Context, id string) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM records WHERE id = ?`, id)
	if err != nil {
		return 0, errors.NewDatabase("Failed to delete record", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, errors.NewDatabase("Failed to get affected rows", err)
	}

	if rowsAffected == 0 {
		return 0, errors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), nil)
	}

	if err := s.UpdateSummary(ctx); err != nil {
		slog.ErrorContext(ctx, "Failed to update summary after record deletion", "record_id", id, "error", err)
	}

	return rowsAffected, nil
}

func (s *RecordService) ExportRecords(ctx context.Context) (*sql.Rows, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.date, r.description, COALESCE(c.name, ''), r.amount, r.type, r.note
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
		ORDER BY r.date DESC
	`)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
