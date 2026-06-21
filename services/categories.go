package services

import (
	"context"
	"log/slog"
	"strings"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

	"github.com/lithammer/shortuuid/v4"
)

func (s *RecordService) CreateCategories(ctx context.Context, categories []models.Category) ([]models.Category, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, errors.NewDatabase("Failed to begin transaction", err)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT INTO categories (id, name, icon, color) VALUES (?, ?, ?, ?)")
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			slog.ErrorContext(ctx, "Failed to rollback transaction", "error", rbErr)
		}
		return nil, errors.NewDatabase("Failed to prepare statement", err)
	}
	defer func() { _ = stmt.Close() }()

	inserted := make([]models.Category, 0, len(categories))
	for _, cat := range categories {
		if cat.Name == "" {
			continue
		}
		catID := shortuuid.New()
		lowerName := strings.ToLower(cat.Name)
		_, err := stmt.ExecContext(ctx, catID, lowerName, cat.Icon, cat.Color)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.ErrorContext(ctx, "Failed to rollback transaction", "error", rbErr)
			}
			return nil, errors.NewDatabase("Failed to insert category", err).WithDetails(map[string]any{
				"categoryName": cat.Name,
			})
		}
		inserted = append(inserted, models.Category{
			ID:    catID,
			Name:  lowerName,
			Icon:  cat.Icon,
			Color: cat.Color,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.NewDatabase("Failed to commit transaction", err)
	}

	return inserted, nil
}

func (s *RecordService) GetCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT "ID", name, icon, color FROM categories ORDER BY name ASC`)
	if err != nil {
		return nil, errors.NewDatabase("Failed to fetch categories", err)
	}
	defer func() { _ = rows.Close() }()

	categories := make([]models.Category, 0)
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Icon, &cat.Color); err != nil {
			slog.WarnContext(ctx, "Failed to scan category row", "error", err)
			continue
		}
		categories = append(categories, cat)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.NewDatabase("Error iterating through categories", err)
	}

	return categories, nil
}

func (s *RecordService) UpdateCategory(ctx context.Context, id string, cat *models.Category) error {
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories WHERE "ID" = ?`, id).Scan(&exists)
	if err != nil {
		return errors.NewDatabase("Failed to check category existence", err)
	}

	if exists == 0 {
		return errors.NewNotFound("Category not found", nil).WithDetails(map[string]any{
			"categoryID": id,
		})
	}

	_, err = s.db.ExecContext(ctx, `UPDATE categories SET name = ?, icon = ?, color = ? WHERE "ID" = ?`,
		cat.Name, cat.Icon, cat.Color, id)
	if err != nil {
		return errors.NewDatabase("Failed to update category", err)
	}

	return nil
}

func (s *RecordService) DeleteCategory(ctx context.Context, id string) error {
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM categories WHERE "ID" = ?`, id).Scan(&exists)
	if err != nil {
		return errors.NewDatabase("Failed to check category existence", err)
	}

	if exists == 0 {
		return errors.NewNotFound("Category not found", nil).WithDetails(map[string]any{
			"categoryID": id,
		})
	}

	var recordCount int
	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM records WHERE "categoryID" = ?`, id).Scan(&recordCount)
	if err != nil {
		return errors.NewDatabase("Failed to check category usage", err)
	}

	if recordCount > 0 {
		return errors.NewConflict("Cannot delete category that has associated records", nil).WithDetails(map[string]any{
			"categoryID":  id,
			"recordCount": recordCount,
		})
	}

	_, err = s.db.ExecContext(ctx, `DELETE FROM categories WHERE "ID" = ?`, id)
	if err != nil {
		return errors.NewDatabase("Failed to delete category", err)
	}

	return nil
}
