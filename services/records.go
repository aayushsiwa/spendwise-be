package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

	"gorm.io/gorm"
)

func (s *RecordService) CreateRecord(ctx context.Context, rec *models.Record) error {
	var category models.Category
	err := s.db.WithContext(ctx).Where("name = ?", strings.ToLower(rec.Category)).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewInvalidInput("Category not found", err).WithDetails(map[string]any{
				"category": rec.Category,
			})
		}
		return apperrors.NewDatabase("Failed to find category", err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return apperrors.NewDatabase("Failed to begin transaction", tx.Error)
	}

	rec.CategoryID = &category.ID
	rec.Balance = 0

	if err := tx.Create(rec).Error; err != nil {
		tx.Rollback()
		return apperrors.NewDatabase("Failed to insert record", err)
	}

	if err = recalculateBalances(ctx, tx); err != nil {
		tx.Rollback()
		return apperrors.NewDatabase("Failed to recalculate balances", err)
	}

	if err = s.updateSummaryTx(ctx, tx); err != nil {
		tx.Rollback()
		slog.ErrorContext(ctx, "Failed to update summary after record creation", "record_id", rec.ID, "error", err)
		return err
	}

	if err = tx.Commit().Error; err != nil {
		return apperrors.NewDatabase("Failed to commit transaction", err)
	}

	return nil
}

func (s *RecordService) GetRecord(ctx context.Context, id string) (*models.Record, error) {
	var rec models.Record
	err := s.db.WithContext(ctx).Table("records").
		Select("records.*, COALESCE(categories.name, '') as category").
		Joins("LEFT JOIN categories ON records.categoryID = categories.ID").
		Where("records.id = ?", id).
		First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), err)
		}
		return nil, apperrors.NewDatabase("Failed to read record", err)
	}
	return &rec, nil
}

func BuildWhereClauseGORM(db *gorm.DB, q *models.QueryParams) *gorm.DB {
	if q.Type != "" {
		db = db.Where("records.type = ?", q.Type)
	}
	if q.Category != "" {
		db = db.Where("categories.name = ?", strings.ToLower(q.Category))
	}
	if q.From != "" {
		db = db.Where("records.date >= ?", q.From)
	}
	if q.To != "" {
		db = db.Where("records.date <= ?", q.To)
	}
	if q.MinAmount != 0 {
		db = db.Where("records.amount >= ?", q.MinAmount)
	}
	if q.MaxAmount != 0 {
		db = db.Where("records.amount <= ?", q.MaxAmount)
	}
	if q.Search != "" {
		db = db.Where("LOWER(records.description) LIKE ?", "%"+strings.ToLower(q.Search)+"%")
	}
	return db
}

func (s *RecordService) GetRecords(ctx context.Context, params *models.QueryParams) ([]models.Record, int, error) {
	offset := (params.Page - 1) * params.Limit

	var records []models.Record
	var totalCount int64

	dbQuery := s.db.WithContext(ctx).Table("records").
		Joins("LEFT JOIN categories ON records.categoryID = categories.ID")

	dbQuery = BuildWhereClauseGORM(dbQuery, params)

	err := dbQuery.Count(&totalCount).Error
	if err != nil {
		return nil, 0, apperrors.NewDatabase("Failed to count records", err)
	}

	err = dbQuery.Select("records.*, COALESCE(categories.name, '') as category").
		Order("records.date DESC").
		Limit(params.Limit).
		Offset(offset).
		Find(&records).Error
	if err != nil {
		return nil, 0, apperrors.NewDatabase("Failed to retrieve records", err)
	}

	return records, int(totalCount), nil
}

func getMonthExpression(dialect string, columnName string) string {
	switch dialect {
	case "postgres":
		return "SUBSTRING(" + columnName + " FROM 1 FOR 7)"
	case "mysql":
		return "SUBSTRING(" + columnName + ", 1, 7)"
	case "sqlite":
		fallthrough
	default:
		return "strftime('%Y-%m', " + columnName + ")"
	}
}

func (s *RecordService) GetGroupedRecords(ctx context.Context, params *models.QueryParams) ([]models.GroupedRecord, error) {
	var groupExpr string
	dialect := s.db.Name()
	switch params.GroupBy {
	case "category":
		groupExpr = "COALESCE(categories.name, '')"
	case "month":
		groupExpr = getMonthExpression(dialect, "records.date")
	default:
		return nil, apperrors.NewInvalidInput("Invalid groupBy value", nil)
	}

	dbQuery := s.db.WithContext(ctx).Table("records").
		Joins("LEFT JOIN categories ON records.categoryID = categories.ID")

	dbQuery = BuildWhereClauseGORM(dbQuery, params)

	var groups []models.GroupedRecord
	err := dbQuery.Select(fmt.Sprintf("%s AS %s, SUM(records.amount) AS total, COUNT(*) AS count", groupExpr, "\"group\"")).
		Group(groupExpr).
		Order("total DESC").
		Scan(&groups).Error
	if err != nil {
		return nil, apperrors.NewDatabase("Failed to retrieve grouped records", err)
	}

	return groups, nil
}

func (s *RecordService) PatchRecord(ctx context.Context, id string, req *models.UpdateRecordRequest) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return apperrors.NewDatabase("Failed to begin transaction", tx.Error)
	}

	var rec models.Record
	err := tx.Where("id = ?", id).First(&rec).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), nil)
		}
		return apperrors.NewDatabase("Failed to find record", err)
	}

	updates := make(map[string]any)
	if req.Date != nil {
		updates["date"] = *req.Date
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.Note != nil {
		updates["note"] = *req.Note
	}
	if req.Category != nil {
		var category models.Category
		err := tx.Where("name = ?", strings.ToLower(*req.Category)).First(&category).Error
		if err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.NewInvalidInput("Category not found", err).WithDetails(map[string]any{
					"category": *req.Category,
				})
			}
			return apperrors.NewDatabase("Failed to find category", err)
		}
		updates["categoryID"] = category.ID
	}

	if len(updates) > 0 {
		err = tx.Model(&models.Record{}).Where("id = ?", id).Updates(updates).Error
		if err != nil {
			tx.Rollback()
			return apperrors.NewDatabase("Failed to update record", err)
		}

		if err = recalculateBalances(ctx, tx); err != nil {
			tx.Rollback()
			return apperrors.NewDatabase("Failed to recalculate balances", err)
		}

		if err = s.updateSummaryTx(ctx, tx); err != nil {
			tx.Rollback()
			slog.ErrorContext(ctx, "Failed to update summary after record update", "record_id", id, "error", err)
			return err
		}
	}

	if err = tx.Commit().Error; err != nil {
		return apperrors.NewDatabase("Failed to commit transaction", err)
	}

	return nil
}

func (s *RecordService) DeleteRecord(ctx context.Context, id string) (int64, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, apperrors.NewDatabase("Failed to begin transaction", tx.Error)
	}

	res := tx.Where("id = ?", id).Delete(&models.Record{})
	if res.Error != nil {
		tx.Rollback()
		return 0, apperrors.NewDatabase("Failed to delete record", res.Error)
	}

	if res.RowsAffected == 0 {
		tx.Rollback()
		return 0, apperrors.NewNotFound(fmt.Sprintf("Record with ID %s not found", id), nil)
	}

	if err := recalculateBalances(ctx, tx); err != nil {
		tx.Rollback()
		return 0, apperrors.NewDatabase("Failed to recalculate balances", err)
	}

	if err := s.updateSummaryTx(ctx, tx); err != nil {
		tx.Rollback()
		slog.ErrorContext(ctx, "Failed to update summary after record deletion", "record_id", id, "error", err)
		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, apperrors.NewDatabase("Failed to commit transaction", err)
	}

	return res.RowsAffected, nil
}

func (s *RecordService) ExportRecords(ctx context.Context) ([]models.Record, error) {
	var records []models.Record
	err := s.db.WithContext(ctx).Table("records").
		Select("records.*, COALESCE(categories.name, '') as category").
		Joins("LEFT JOIN categories ON records.categoryID = categories.ID").
		Order("records.date DESC").
		Find(&records).Error
	if err != nil {
		return nil, apperrors.NewDatabase("Failed to export records", err)
	}
	return records, nil
}
