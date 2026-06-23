package services

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/utils"

	"gorm.io/gorm"
)

func (s *RecordService) updateSummaryTx(ctx context.Context, tx *gorm.DB) (err error) {
	slog.InfoContext(ctx, "Updating summary...")

	if err = tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.SummaryDB{}).Error; err != nil {
		return errors.NewDatabase("Failed to clear summary", err)
	}
	if err = tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.SummaryDetailDB{}).Error; err != nil {
		return errors.NewDatabase("Failed to clear summary_details", err)
	}

	dialect := tx.Name()
	monthExpr := getMonthExpression(dialect, "date")

	var count int64
	if err = tx.Model(&models.Record{}).Count(&count).Error; err != nil {
		return errors.NewDatabase("Failed to count records for min month", err)
	}
	if count == 0 {
		slog.InfoContext(ctx, "No records found, summary will be empty")
		return nil
	}

	var minMonthStr string
	err = tx.Model(&models.Record{}).Select("MIN(" + monthExpr + ")").Row().Scan(&minMonthStr)
	if err != nil {
		return errors.NewDatabase("Failed to get min month", err)
	}

	if minMonthStr == "" {
		slog.InfoContext(ctx, "No records found, summary will be empty")
		return nil
	}

	maxMonth := time.Now().Format("2006-01")

	type MonthAgg struct {
		Month        string
		TotalIncome  float64
		TotalExpense float64
	}
	var aggs []MonthAgg
	err = tx.Model(&models.Record{}).
		Select(monthExpr + " AS month, SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) AS total_income, SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) AS total_expense").
		Group("month").
		Order("month ASC").
		Scan(&aggs).Error
	if err != nil {
		return errors.NewDatabase("Failed to aggregate records", err)
	}

	type monthData struct {
		income  float64
		expense float64
	}
	data := make(map[string]monthData)
	for _, agg := range aggs {
		data[agg.Month] = monthData{agg.TotalIncome, agg.TotalExpense}
	}

	openingBalance := 0.0
	var summariesToInsert []models.SummaryDB
	for m := minMonthStr; m <= maxMonth; m = utils.NextMonth(m) {
		d := data[m]
		net := d.income - d.expense
		closing := openingBalance + net

		summariesToInsert = append(summariesToInsert, models.SummaryDB{
			Month:          m,
			TotalIncome:    d.income,
			TotalExpense:   d.expense,
			OpeningBalance: openingBalance,
			NetBalance:     net,
			ClosingBalance: closing,
		})
		openingBalance = closing
	}

	if len(summariesToInsert) > 0 {
		if err = tx.Create(&summariesToInsert).Error; err != nil {
			return errors.NewDatabase("Failed to insert summary data", err)
		}
	}

	type DetailAgg struct {
		Month        string
		Type         string
		CategoryID   string
		CategoryName string
		Amount       float64
	}
	var details []DetailAgg

	coalesceID := "COALESCE(categories.ID, '')"
	coalesceName := "COALESCE(categories.name, 'uncategorized')"
	monthExprWithAlias := getMonthExpression(dialect, "records.date")

	err = tx.Table("records").
		Select(monthExprWithAlias+" AS month, records.type AS type, "+coalesceID+" AS category_id, "+coalesceName+" AS category_name, SUM(records.amount) AS amount").
		Joins("LEFT JOIN categories ON records.categoryID = categories.ID").
		Where("records.type IN ?", []string{"income", "expense", "transfer"}).
		Group("month, records.type, " + coalesceID + ", " + coalesceName).
		Scan(&details).Error
	if err != nil {
		return errors.NewDatabase("Failed to fetch aggregate category details", err)
	}

	var detailsToInsert []models.SummaryDetailDB
	for _, det := range details {
		detailsToInsert = append(detailsToInsert, models.SummaryDetailDB{
			Month:        det.Month,
			Type:         det.Type,
			CategoryID:   det.CategoryID,
			CategoryName: det.CategoryName,
			Amount:       det.Amount,
		})
	}

	if len(detailsToInsert) > 0 {
		if err = tx.Create(&detailsToInsert).Error; err != nil {
			return errors.NewDatabase("Failed to insert summary details", err)
		}
	}

	slog.InfoContext(ctx, "Summary updated successfully")
	return nil
}

func (s *RecordService) UpdateSummary(ctx context.Context) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return errors.NewDatabase("Failed to begin transaction", tx.Error)
	}

	if err := s.updateSummaryTx(ctx, tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return errors.NewDatabase("Failed to commit summary transaction", err)
	}
	return nil
}

func (s *RecordService) GetSummary(ctx context.Context, from, to, categoryFilter, typeFilter string) (*models.Summary, error) {
	var totals struct {
		TotalIncome  float64
		TotalExpense float64
	}
	err := s.db.WithContext(ctx).Model(&models.Record{}).
		Select("COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS total_income, COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS total_expense").
		Where("date >= ? AND date <= ?", from, to).
		Scan(&totals).Error
	if err != nil {
		return nil, errors.NewDatabase("Failed to compute totals", err)
	}

	var openingBalance float64
	err = s.db.WithContext(ctx).Model(&models.Record{}).
		Select("COALESCE(SUM(CASE WHEN type = 'income' THEN amount WHEN type = 'expense' THEN -amount ELSE 0 END), 0)").
		Where("date < ?", from).
		Scan(&openingBalance).Error
	if err != nil {
		return nil, errors.NewDatabase("Failed to compute opening balance", err)
	}

	netInRange := totals.TotalIncome - totals.TotalExpense
	closingBalance := openingBalance + netInRange

	dbQuery := s.db.WithContext(ctx).Table("records").
		Select("COALESCE(categories.ID, '') as category_id, COALESCE(categories.name, '') as category_name, records.type as type, SUM(records.amount) as amount").
		Joins("LEFT JOIN categories ON records.categoryID = categories.ID").
		Where("records.date >= ? AND records.date <= ?", from, to)

	if categoryFilter != "" {
		dbQuery = dbQuery.Where("LOWER(categories.name) = ?", strings.ToLower(categoryFilter))
	}
	if typeFilter != "" {
		dbQuery = dbQuery.Where("records.type = ?", typeFilter)
	}

	type SummaryDetailAgg struct {
		CategoryID   string
		CategoryName string
		Type         string
		Amount       float64
	}
	var details []SummaryDetailAgg

	err = dbQuery.Group("categories.ID, categories.name, records.type").
		Order("records.type, SUM(records.amount) DESC").
		Scan(&details).Error
	if err != nil {
		return nil, errors.NewDatabase("Failed to fetch category breakdown", err)
	}

	incomes := make([]models.CategoryDetail, 0)
	expenses := make([]models.CategoryDetail, 0)
	for _, det := range details {
		cd := models.CategoryDetail{
			CategoryID: det.CategoryID,
			Category:   det.CategoryName,
			Amount:     det.Amount,
		}
		switch det.Type {
		case "income":
			incomes = append(incomes, cd)
		case "expense":
			expenses = append(expenses, cd)
		}
	}

	return &models.Summary{
		Expenses:     expenses,
		Incomes:      incomes,
		Net:          netInRange,
		Opening:      openingBalance,
		Closing:      closingBalance,
		TotalExpense: totals.TotalExpense,
		TotalIncome:  totals.TotalIncome,
	}, nil
}
