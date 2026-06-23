package services

import (
	"context"
	"fmt"
	"time"

	"aayushsiwa/expense-tracker/models"

	"gorm.io/gorm"
)

func (s *RecordService) CreateBudget(ctx context.Context, budget *models.Budget) error {
	return s.db.WithContext(ctx).Create(budget).Error
}

func (s *RecordService) GetBudgets(ctx context.Context, month, year int) ([]models.Budget, error) {
	var budgets []models.Budget
	err := s.db.WithContext(ctx).
		Select("budgets.*, COALESCE(categories.name, '') as category").
		Joins("LEFT JOIN categories ON budgets.categoryID = categories.ID").
		Where("budgets.month = ? AND budgets.year = ?", month, year).
		Order("categories.name ASC").
		Find(&budgets).Error
	return budgets, err
}

type spentResult struct {
	CategoryID string
	Spent      float64
}

func (s *RecordService) GetBudgetProgress(ctx context.Context, month, year int) ([]models.BudgetProgress, error) {
	budgets, err := s.GetBudgets(ctx, month, year)
	if err != nil {
		return nil, err
	}
	if len(budgets) == 0 {
		return nil, nil
	}

	startDate := fmt.Sprintf("%04d-%02d-01", year, month)
	endDate := fmt.Sprintf("%04d-%02d-%02d", year, month, daysInMonth(year, time.Month(month)))

	var spent []spentResult
	err = s.db.WithContext(ctx).
		Model(&models.Record{}).
		Select("categoryID, COALESCE(SUM(amount), 0) AS spent").
		Where("type = ? AND date >= ? AND date <= ?", models.Expense, startDate, endDate).
		Group("categoryID").
		Scan(&spent).Error
	if err != nil {
		return nil, err
	}

	spentMap := make(map[string]float64, len(spent))
	for _, s := range spent {
		spentMap[s.CategoryID] = s.Spent
	}

	progress := make([]models.BudgetProgress, 0, len(budgets))
	for _, b := range budgets {
		sp := spentMap[b.CategoryID]
		pct := 0.0
		if b.Amount > 0 {
			pct = (sp / b.Amount) * 100
		}
		progress = append(progress, models.BudgetProgress{
			Budget:     b,
			Spent:      sp,
			Percentage: pct,
		})
	}

	return progress, nil
}

func (s *RecordService) UpdateBudget(ctx context.Context, id string, amount float64) error {
	result := s.db.WithContext(ctx).Model(&models.Budget{}).Where("ID = ?", id).Update("amount", amount)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *RecordService) DeleteBudget(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Where("ID = ?", id).Delete(&models.Budget{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func daysInMonth(year int, m time.Month) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
