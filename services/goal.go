package services

import (
	"context"
	"errors"
	"strings"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

	"github.com/lithammer/shortuuid/v4"
	"gorm.io/gorm"
)

func (s *RecordService) CreateGoal(ctx context.Context, goal *models.Goal) error {
	goal.ID = shortuuid.New()
	goal.CurrentAmount = 0
	if goal.Status == "" {
		goal.Status = models.GoalActive
	}

	category, catErr := s.resolveCategory(ctx, s.db, goal.Category)
	if catErr != nil {
		return catErr
	}
	if category != nil {
		goal.CategoryID = &category.ID
	}

	return s.db.WithContext(ctx).Create(goal).Error
}

func (s *RecordService) GetGoals(ctx context.Context) ([]models.Goal, error) {
	var goals []models.Goal
	err := s.db.WithContext(ctx).Table("goals").
		Select("goals.*, COALESCE(categories.name, '') as category").
		Joins("LEFT JOIN categories ON goals.categoryID = categories.ID").
		Order("goals.status ASC, goals.name ASC").
		Scan(&goals).Error
	if err != nil {
		return nil, apperrors.NewDatabase("Failed to fetch goals", err)
	}
	return goals, nil
}

func (s *RecordService) GetGoal(ctx context.Context, id string) (*models.Goal, error) {
	var goal models.Goal
	err := s.db.WithContext(ctx).Table("goals").
		Select("goals.*, COALESCE(categories.name, '') as category").
		Joins("LEFT JOIN categories ON goals.categoryID = categories.ID").
		Where("goals.id = ?", id).
		First(&goal).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFound("Goal not found", nil)
		}
		return nil, apperrors.NewDatabase("Failed to fetch goal", err)
	}
	return &goal, nil
}

func (s *RecordService) UpdateGoal(ctx context.Context, id string, req *models.UpdateGoalRequest) error {
	updates := make(map[string]any)

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.TargetAmount != nil {
		updates["targetAmount"] = *req.TargetAmount
	}
	if req.CurrentAmount != nil {
		updates["currentAmount"] = *req.CurrentAmount
	}
	if req.TargetDate != nil {
		updates["targetDate"] = *req.TargetDate
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.MonthlyContribution != nil {
		updates["monthlyContribution"] = *req.MonthlyContribution
	}
	if req.Category != nil {
		category, catErr := s.resolveCategory(ctx, s.db, *req.Category)
		if catErr != nil {
			return catErr
		}
		if category != nil {
			updates["categoryID"] = category.ID
		} else {
			updates["categoryID"] = nil
		}
	}

	result := s.db.WithContext(ctx).Model(&models.Goal{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return apperrors.NewDatabase("Failed to update goal", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewNotFound("Goal not found", nil)
	}
	return nil
}

func (s *RecordService) DeleteGoal(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Goal{})
	if result.Error != nil {
		return apperrors.NewDatabase("Failed to delete goal", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewNotFound("Goal not found", nil)
	}
	return nil
}

func (s *RecordService) AddGoalProgress(ctx context.Context, id string, amount float64) error {
	result := s.db.WithContext(ctx).Model(&models.Goal{}).
		Where("id = ?", id).
		UpdateColumn("currentAmount", gorm.Expr("currentAmount + ?", amount))
	if result.Error != nil {
		return apperrors.NewDatabase("Failed to add goal progress", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperrors.NewNotFound("Goal not found", nil)
	}
	return nil
}

func (s *RecordService) resolveCategory(ctx context.Context, db *gorm.DB, categoryName string) (*models.Category, error) {
	if categoryName == "" {
		return nil, nil
	}
	lowerName := strings.ToLower(strings.TrimSpace(categoryName))
	var category models.Category
	err := db.WithContext(ctx).Where("LOWER(name) = ?", lowerName).First(&category).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewInvalidInput("Category not found", err).WithDetails(map[string]any{
				"category": categoryName,
			})
		}
		return nil, apperrors.NewDatabase("Failed to find category", err)
	}
	return &category, nil
}
