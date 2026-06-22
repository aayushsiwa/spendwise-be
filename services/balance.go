package services

import (
	"context"
	"log/slog"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

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

// recalculateBalances updates all records' balance values to reflect cumulative running totals computed in chronological order.
func recalculateBalances(ctx context.Context, tx *gorm.DB) error {
	var records []models.Record
	err := tx.Select("id, date, amount, type").Order("date ASC, id ASC").Find(&records).Error
	if err != nil {
		return err
	}

	var runningBalance float64
	for i := range records {
		if records[i].Type == models.Income {
			runningBalance += records[i].Amount
		} else {
			runningBalance -= records[i].Amount
		}

		err = tx.Model(&models.Record{}).Where("id = ?", records[i].ID).Update("balance", runningBalance).Error
		if err != nil {
			return err
		}
	}

	return nil
}
