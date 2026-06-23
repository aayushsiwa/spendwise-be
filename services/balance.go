package services

import (
	"context"
	"log/slog"

	"aayushsiwa/expense-tracker/errors"

	"gorm.io/gorm"
)

func (s *RecordService) RefreshBalances(ctx context.Context) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.NewDatabase("Failed to start transaction", tx.Error)
	}

	if err := recalculateBalances(ctx, tx); err != nil {
		tx.Rollback()
		slog.ErrorContext(ctx, "Failed to recalculate balances", "error", err)
		return errors.NewDatabase("Failed to recalculate balances", err)
	}

	if err := tx.Commit().Error; err != nil {
		return errors.NewDatabase("Failed to commit transaction", err)
	}

	return nil
}

// recalculateBalances updates all records' balance values to cumulative running totals in a single query.
func recalculateBalances(ctx context.Context, tx *gorm.DB) error {
	// Get the database dialect name
	dialect := tx.Dialector.Name()

	var query string

	// MySQL 8.0+ does not allow the target table to be referenced in a subquery of the SET clause,
	// even through a CTE. Use JOIN-based UPDATE instead.
	if dialect == "mysql" {
		query = `
			UPDATE records r
			JOIN (
				SELECT id,
					SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END)
						OVER (ORDER BY date ASC, id ASC) AS running_balance
				FROM records
			) sub ON r.id = sub.id
			SET r.balance = sub.running_balance
		`
	} else {
		// PostgreSQL and SQLite support CTE-based UPDATE with subqueries
		query = `
			WITH ordered_balances AS (
				SELECT id,
					SUM(CASE WHEN type = 'income' THEN amount ELSE -amount END)
						OVER (ORDER BY date ASC, id ASC) AS running_balance
				FROM records
			)
			UPDATE records
			SET balance = (
				SELECT running_balance
				FROM ordered_balances
				WHERE ordered_balances.id = records.id
			)
		`
	}

	return tx.WithContext(ctx).Exec(query).Error
}
