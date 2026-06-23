package services

import (
	"context"
	"strings"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

	"github.com/lithammer/shortuuid/v4"
)

func (s *RecordService) CreateCategories(ctx context.Context, categories []models.Category) ([]models.Category, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, errors.NewDatabase("Failed to begin transaction", tx.Error)
	}

	inserted := make([]models.Category, 0, len(categories))
	for _, cat := range categories {
		if cat.Name == "" {
			continue
		}
		catID := shortuuid.New()
		lowerName := strings.ToLower(cat.Name)
		newCat := models.Category{
			ID:    catID,
			Name:  lowerName,
			Icon:  cat.Icon,
			Color: cat.Color,
		}

		if err := tx.Create(&newCat).Error; err != nil {
			tx.Rollback()
			return nil, errors.NewDatabase("Failed to insert category", err).WithDetails(map[string]any{
				"categoryName": cat.Name,
			})
		}
		inserted = append(inserted, newCat)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.NewDatabase("Failed to commit transaction", err)
	}

	return inserted, nil
}

func (s *RecordService) GetCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	err := s.db.WithContext(ctx).Order("name ASC").Find(&categories).Error
	if err != nil {
		return nil, errors.NewDatabase("Failed to fetch categories", err)
	}
	return categories, nil
}

func (s *RecordService) UpdateCategory(ctx context.Context, id string, cat *models.Category) error {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Category{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return errors.NewDatabase("Failed to check category existence", err)
	}

	if count == 0 {
		return errors.NewNotFound("Category not found", nil).WithDetails(map[string]any{
			"categoryID": id,
		})
	}

	err = s.db.WithContext(ctx).Model(&models.Category{}).Where("id = ?", id).Updates(map[string]any{
		"name":  strings.ToLower(cat.Name),
		"icon":  cat.Icon,
		"color": cat.Color,
	}).Error
	if err != nil {
		return errors.NewDatabase("Failed to update category", err)
	}

	return nil
}

func (s *RecordService) DeleteCategory(ctx context.Context, id string) error {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.Category{}).Where("id = ?", id).Count(&count).Error
	if err != nil {
		return errors.NewDatabase("Failed to check category existence", err)
	}

	if count == 0 {
		return errors.NewNotFound("Category not found", nil).WithDetails(map[string]any{
			"categoryID": id,
		})
	}

	var recordCount int64
	err = s.db.WithContext(ctx).Model(&models.Record{}).Where("categoryID = ?", id).Count(&recordCount).Error
	if err != nil {
		return errors.NewDatabase("Failed to check category usage", err)
	}

	if recordCount > 0 {
		return errors.NewConflict("Cannot delete category that has associated records", nil).WithDetails(map[string]any{
			"categoryID":  id,
			"recordCount": recordCount,
		})
	}

	err = s.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Category{}).Error
	if err != nil {
		return errors.NewDatabase("Failed to delete category", err)
	}

	return nil
}
