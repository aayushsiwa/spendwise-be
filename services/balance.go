package services

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"

	"aayushsiwa/expense-tracker/errors"
)

func (s *RecordService) RefreshBalances(ctx context.Context) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewDatabase("Failed to start transaction", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = recalculateBalances(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to recalculate balances", "error", err)
		return errors.NewDatabase("Failed to recalculate balances", err)
	}

	if err = tx.Commit(); err != nil {
		return errors.NewDatabase("Failed to commit transaction", err)
	}

	return nil
}

func recalculateBalances(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, date, amount, type
		FROM records
		ORDER BY date ASC, id ASC
	`)
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			slog.ErrorContext(ctx, "Error closing rows", "err", err)
		}
	}(rows)

	var (
		runningBalance float64
		batchSize      = 100
		ids            []string
		balances       []float64
	)

	updateBatch := func() error {
		if len(ids) == 0 {
			return nil
		}
		var query strings.Builder
		query.WriteString("UPDATE records SET balance = CASE id ")
		var args []any
		for i := range ids {
			query.WriteString("WHEN ? THEN ? ")
			args = append(args, ids[i], balances[i])
		}
		query.WriteString("END WHERE id IN (")
		for i := range ids {
			if i > 0 {
				query.WriteString(",")
			}
			query.WriteString("?")
			args = append(args, ids[i])
		}
		query.WriteString(")")
		_, err := tx.ExecContext(ctx, query.String(), args...)
		return err
	}

	for rows.Next() {
		var id string
		var date string
		var amount float64
		var recordType string

		if err := rows.Scan(&id, &date, &amount, &recordType); err != nil {
			return err
		}

		if recordType == "income" {
			runningBalance += amount
		} else {
			runningBalance -= amount
		}

		ids = append(ids, id)
		balances = append(balances, runningBalance)

		if len(ids) >= batchSize {
			if err := updateBatch(); err != nil {
				return err
			}
			ids = nil
			balances = nil
		}
	}

	if err := updateBatch(); err != nil {
		return err
	}

	return nil
}
