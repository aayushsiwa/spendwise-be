package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

	"github.com/lithammer/shortuuid/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	gormDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	err = gormDB.AutoMigrate(
		&models.Category{},
		&models.Record{},
		&models.SummaryDB{},
		&models.SummaryDetailDB{},
		&models.Budget{},
		&models.Goal{},
	)
	if err != nil {
		t.Fatalf("failed to auto-migrate schema: %v", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatalf("failed to get underlying db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	return gormDB
}

func isAppErrorType(err error, typ string) bool {
	var appErr *apperrors.AppError
	return errors.As(err, &appErr) && appErr.Type == typ
}

func TestCreateCategories(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	created, err := svc.CreateCategories(context.Background(), []models.Category{
		{Name: "Food", Icon: "🍕", Color: "#ff0000"},
		{Name: "Transport", Icon: "🚗", Color: "#00ff00"},
	})
	if err != nil {
		t.Fatalf("CreateCategories failed: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(created))
	}
	for _, c := range created {
		if c.ID == "" {
			t.Error("expected non-empty ID")
		}
	}
	if created[0].Name != "food" || created[1].Name != "transport" {
		t.Errorf("names not lowercased: %+v", created)
	}
}

func TestGetCategories(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	cats, err := svc.GetCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 0 {
		t.Fatalf("expected 0, got %d", len(cats))
	}

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}
	cats, err = svc.GetCategories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(cats) != 1 || cats[0].Name != "food" {
		t.Fatalf("unexpected: %+v", cats)
	}
}

func TestCreateRecord(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}

	rec := &models.Record{
		ID:          shortuuid.New(),
		Date:        "2024-01-15",
		Description: "Groceries",
		Category:    "Food",
		Amount:      50.0,
		Type:        models.Expense,
		Note:        "weekly shopping",
	}
	if err := svc.CreateRecord(context.Background(), rec); err != nil {
		t.Fatalf("CreateRecord failed: %v", err)
	}
}

func TestGetRecord(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}

	rec := &models.Record{
		ID:          shortuuid.New(),
		Date:        "2024-01-15",
		Description: "Groceries",
		Category:    "Food",
		Amount:      50.0,
		Type:        models.Expense,
	}
	if err := svc.CreateRecord(context.Background(), rec); err != nil {
		t.Fatal(err)
	}

	got, err := svc.GetRecord(context.Background(), rec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != rec.ID || got.Description != "Groceries" {
		t.Errorf("unexpected record: %+v", got)
	}
	if got.Category != "food" {
		t.Errorf("category = %q, want food", got.Category)
	}
	if got.Balance != -50.0 {
		t.Errorf("balance = %f, want -50", got.Balance)
	}

	_, err = svc.GetRecord(context.Background(), "nonexistent")
	if !isAppErrorType(err, "not_found") {
		t.Errorf("expected not_found, got %v", err)
	}
}

func TestDeleteRecord(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)
	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}

	rec := &models.Record{
		ID:          shortuuid.New(),
		Date:        "2024-01-15",
		Description: "Groceries",
		Category:    "Food",
		Amount:      50.0,
		Type:        models.Expense,
	}
	if err := svc.CreateRecord(context.Background(), rec); err != nil {
		t.Fatal(err)
	}

	n, err := svc.DeleteRecord(context.Background(), rec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("rows affected = %d, want 1", n)
	}

	_, err = svc.GetRecord(context.Background(), rec.ID)
	if !isAppErrorType(err, "not_found") {
		t.Errorf("expected not_found after delete, got %v", err)
	}
}

func TestDeleteCategory(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	cats, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}})
	if err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}
	catID := cats[0].ID

	if err = svc.DeleteCategory(context.Background(), catID); err != nil {
		t.Fatalf("DeleteCategory failed: %v", err)
	}

	err = svc.DeleteCategory(context.Background(), catID)
	if !isAppErrorType(err, "not_found") {
		t.Errorf("expected not_found for second delete, got %v", err)
	}
}

func TestDeleteCategory_Conflict(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	cats, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}})
	if err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}
	catID := cats[0].ID

	if err := svc.CreateRecord(context.Background(), &models.Record{
		ID: shortuuid.New(), Date: "2024-01-15", Description: "Groceries", Category: "Food",
		Amount: 50.0, Type: models.Expense,
	}); err != nil {
		t.Fatalf("CreateRecord setup failed: %v", err)
	}

	err = svc.DeleteCategory(context.Background(), catID)
	if !isAppErrorType(err, "conflict") {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestGetSummary(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{
		{Name: "Salary"}, {Name: "Food"}, {Name: "Rent"},
	}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}

	for _, rec := range []*models.Record{
		{ID: shortuuid.New(), Date: "2024-01-01", Description: "Salary Jan", Category: "Salary", Amount: 5000, Type: models.Income},
		{ID: shortuuid.New(), Date: "2024-01-05", Description: "Groceries", Category: "Food", Amount: 200, Type: models.Expense},
		{ID: shortuuid.New(), Date: "2024-01-10", Description: "Rent", Category: "Rent", Amount: 1000, Type: models.Expense},
	} {
		if err := svc.CreateRecord(context.Background(), rec); err != nil {
			t.Fatalf("CreateRecord failed: %v", err)
		}
	}

	summary, err := svc.GetSummary(context.Background(), "2024-01-01", "2024-01-31", "", "")
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if summary.TotalIncome != 5000 {
		t.Errorf("TotalIncome = %f, want 5000", summary.TotalIncome)
	}
	if summary.TotalExpense != 1200 {
		t.Errorf("TotalExpense = %f, want 1200", summary.TotalExpense)
	}
	if summary.Net != 3800 {
		t.Errorf("Net = %f, want 3800", summary.Net)
	}
	if summary.Closing != 3800 {
		t.Errorf("Closing = %f, want 3800", summary.Closing)
	}
	if len(summary.Incomes)+len(summary.Expenses) != 3 {
		t.Errorf("expected 3 detail rows, got incomes=%d expenses=%d",
			len(summary.Incomes), len(summary.Expenses))
	}
}

func TestGetSummary_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{
		{Name: "Salary"}, {Name: "Food"}, {Name: "Rent"},
	}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}

	for _, rec := range []*models.Record{
		{ID: shortuuid.New(), Date: "2024-01-01", Description: "Salary Jan", Category: "Salary", Amount: 5000, Type: models.Income},
		{ID: shortuuid.New(), Date: "2024-01-05", Description: "Groceries", Category: "Food", Amount: 200, Type: models.Expense},
		{ID: shortuuid.New(), Date: "2024-01-10", Description: "Rent", Category: "Rent", Amount: 1000, Type: models.Expense},
	} {
		if err := svc.CreateRecord(context.Background(), rec); err != nil {
			t.Fatalf("CreateRecord failed: %v", err)
		}
	}

	summary, err := svc.GetSummary(context.Background(), "2024-01-01", "2024-01-31", "Food", "")
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	// totals are unfiltered at the SQL level
	if summary.TotalIncome != 5000 {
		t.Errorf("TotalIncome = %f, want 5000", summary.TotalIncome)
	}
	if summary.TotalExpense != 1200 {
		t.Errorf("TotalExpense = %f, want 1200", summary.TotalExpense)
	}
	if len(summary.Expenses) != 1 || summary.Expenses[0].Category != "food" {
		t.Errorf("expected 1 Food expense detail, got %+v", summary.Expenses)
	}
}

func TestGetSummary_WithTypeFilter(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{
		{Name: "Salary"}, {Name: "Food"},
	}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}

	for _, rec := range []*models.Record{
		{ID: shortuuid.New(), Date: "2024-01-01", Description: "Salary Jan", Category: "Salary", Amount: 5000, Type: models.Income},
		{ID: shortuuid.New(), Date: "2024-01-05", Description: "Groceries", Category: "Food", Amount: 200, Type: models.Expense},
	} {
		if err := svc.CreateRecord(context.Background(), rec); err != nil {
			t.Fatalf("CreateRecord failed: %v", err)
		}
	}

	summary, err := svc.GetSummary(context.Background(), "2024-01-01", "2024-01-31", "", "expense")
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}
	if len(summary.Incomes) != 0 {
		t.Errorf("expected 0 income details, got %d", len(summary.Incomes))
	}
	if len(summary.Expenses) != 1 || summary.Expenses[0].Category != "food" {
		t.Errorf("expected 1 Food expense detail, got %+v", summary.Expenses)
	}
}

func TestGetSummary_DbError(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetSummary(ctx, "2024-01-01", "2024-01-31", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, context.Canceled) && !isAppErrorType(err, "database") {
		t.Errorf("expected context.Canceled or database AppError, got %v", err)
	}
}

func TestUpdateSummary(t *testing.T) {
	t.Run("empty database", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		err := svc.UpdateSummary(context.Background())
		if err != nil {
			t.Fatalf("UpdateSummary failed: %v", err)
		}

		// Verify summary is empty
		var count int64
		err = db.Table("summary").Count(&count).Error
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("expected 0 summary rows, got %d", count)
		}
	})

	t.Run("with records", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		if _, err := svc.CreateCategories(context.Background(), []models.Category{
			{Name: "Salary"}, {Name: "Food"},
		}); err != nil {
			t.Fatalf("CreateCategories setup failed: %v", err)
		}

		for _, rec := range []*models.Record{
			{ID: shortuuid.New(), Date: "2024-01-01", Description: "Salary Jan", Category: "Salary", Amount: 5000, Type: models.Income},
			{ID: shortuuid.New(), Date: "2024-01-05", Description: "Groceries", Category: "Food", Amount: 200, Type: models.Expense},
		} {
			if err := svc.CreateRecord(context.Background(), rec); err != nil {
				t.Fatalf("CreateRecord failed: %v", err)
			}
		}

		err := svc.UpdateSummary(context.Background())
		if err != nil {
			t.Fatalf("UpdateSummary failed: %v", err)
		}

		// Verify summary rows
		var count int64
		err = db.Table("summary").Count(&count).Error
		if err != nil {
			t.Fatal(err)
		}
		if count == 0 {
			t.Errorf("expected summary rows, got 0")
		}

		// Verify summary details rows
		err = db.Table("summary_details").Count(&count).Error
		if err != nil {
			t.Fatal(err)
		}
		if count == 0 {
			t.Errorf("expected summary details rows, got 0")
		}
	})
}

func TestUpdateSummary_DbError(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := svc.UpdateSummary(ctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, context.Canceled) && !isAppErrorType(err, "database") {
		t.Errorf("expected context.Canceled or database AppError, got %v", err)
	}
}

func TestUpdateSummary_Errors(t *testing.T) {
	t.Run("failed to clear summary", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		db.Exec("DROP TABLE summary")

		err := svc.UpdateSummary(context.Background())
		if err == nil {
			t.Fatal("expected error when summary table is missing")
		}
		if !isAppErrorType(err, "database_error") {
			t.Errorf("expected database error for summary clear failure, got %T: %v", err, err)
		}
	})

	t.Run("failed to clear summary_details", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		db.Exec("DROP TABLE summary_details")

		err := svc.UpdateSummary(context.Background())
		if err == nil {
			t.Fatal("error when summary_details table is missing")
		}
		if !isAppErrorType(err, "database_error") {
			t.Errorf("expected database error for summary_details clear failure, got %T: %v", err, err)
		}
	})

	t.Run("failed to get min month", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		db.Exec("DROP TABLE records")

		err := svc.UpdateSummary(context.Background())
		if err == nil {
			t.Fatal("expected error when records table is missing")
		}
		if !isAppErrorType(err, "database_error") {
			t.Errorf("expected database error for records query failure, got %T: %v", err, err)
		}
	})
}

func TestUpdateSummary_InsertErrors(t *testing.T) {
	t.Run("failed to insert summary", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		// Create a record so minMonth is valid and it proceeds to insert
		if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}}); err != nil {
			t.Fatal(err)
		}
		if err := svc.CreateRecord(context.Background(), &models.Record{
			ID: shortuuid.New(), Date: "2024-01-01", Description: "Groceries", Category: "Food", Amount: 200, Type: models.Expense,
		}); err != nil {
			t.Fatal(err)
		}

		// Recreate summary with failing check constraint
		if err := db.Exec("DROP TABLE summary").Error; err != nil {
			t.Fatalf("failed to drop summary table: %v", err)
		}
		if err := db.Exec("CREATE TABLE summary (month TEXT PRIMARY KEY CHECK (month = 'invalid'))").Error; err != nil {
			t.Fatalf("failed to create mock summary table: %v", err)
		}

		err := svc.UpdateSummary(context.Background())
		if err == nil {
			t.Error("expected error due to CHECK constraint failure on summary table")
		}
	})

	t.Run("failed to insert summary_details", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		// Create a record so minMonth is valid and it proceeds to insert
		if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}}); err != nil {
			t.Fatal(err)
		}
		if err := svc.CreateRecord(context.Background(), &models.Record{
			ID: shortuuid.New(), Date: "2024-01-01", Description: "Groceries", Category: "Food", Amount: 200, Type: models.Expense,
		}); err != nil {
			t.Fatal(err)
		}

		// Recreate summary_details with failing check constraint
		if err := db.Exec("DROP TABLE summary_details").Error; err != nil {
			t.Fatalf("failed to drop summary_details table: %v", err)
		}
		if err := db.Exec("CREATE TABLE summary_details (month TEXT, type TEXT, categoryID TEXT, categoryName TEXT, amount REAL, PRIMARY KEY (month, type, categoryID), CHECK (amount < 0))").Error; err != nil {
			t.Fatalf("failed to create mock summary_details table: %v", err)
		}

		err := svc.UpdateSummary(context.Background())
		if err == nil {
			t.Error("expected error due to CHECK constraint failure on summary_details table")
		}
	})
}

func TestUpdateCategory(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	cats, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Old"}})
	if err != nil {
		t.Fatal(err)
	}

	err = svc.UpdateCategory(context.Background(), cats[0].ID, &models.Category{Name: "New", Icon: "✨", Color: "#abcdef"})
	if err != nil {
		t.Fatalf("failed to update category: %v", err)
	}

	var updated models.Category
	if err = db.Where("id = ?", cats[0].ID).First(&updated).Error; err != nil {
		t.Fatal(err)
	}
	if updated.Name != "new" || updated.Icon != "✨" || updated.Color != "#abcdef" {
		t.Errorf("category not updated properly: %+v", updated)
	}

	// Update nonexistent
	err = svc.UpdateCategory(context.Background(), "nonexistent", &models.Category{Name: "No"})
	if err == nil {
		t.Error("expected error updating nonexistent category")
	}
}

func TestRefreshBalances(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}}); err != nil {
		t.Fatal(err)
	}

	r1 := models.Record{ID: "r1", Date: "2024-01-01", Description: "R1", Category: "Food", Amount: 100, Type: models.Income}
	r2 := models.Record{ID: "r2", Date: "2024-01-02", Description: "R2", Category: "Food", Amount: 40, Type: models.Expense}
	if err := svc.CreateRecord(context.Background(), &r1); err != nil {
		t.Fatal(err)
	}
	if err := svc.CreateRecord(context.Background(), &r2); err != nil {
		t.Fatal(err)
	}

	// Manually mess up balances
	db.Model(&models.Record{}).Where("1 = 1").Update("balance", 0)

	err := svc.RefreshBalances(context.Background())
	if err != nil {
		t.Fatalf("failed to refresh balances: %v", err)
	}

	var dbR1, dbR2 models.Record
	db.Where("id = ?", "r1").First(&dbR1)
	db.Where("id = ?", "r2").First(&dbR2)

	if dbR1.Balance != 100 || dbR2.Balance != 60 {
		t.Errorf("balances not refreshed correctly: r1=%f, r2=%f", dbR1.Balance, dbR2.Balance)
	}
}

func TestGetRecordsAndFilters(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}, {Name: "Salary"}}); err != nil {
		t.Fatal(err)
	}

	_ = svc.CreateRecord(context.Background(), &models.Record{ID: "r1", Date: "2024-01-01", Description: "Lunch", Category: "Food", Amount: 20, Type: models.Expense})
	_ = svc.CreateRecord(context.Background(), &models.Record{ID: "r2", Date: "2024-01-02", Description: "Payday", Category: "Salary", Amount: 1000, Type: models.Income})

	records, total, err := svc.GetRecords(context.Background(), &models.QueryParams{
		PaginationFilterParams: models.PaginationFilterParams{Page: 1, Limit: 10},
		Type:                   "expense",
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(records) != 1 || records[0].ID != "r1" {
		t.Errorf("unexpected getrecords result: total=%d, len=%d", total, len(records))
	}

	// Query with min/max amount and search
	records, total, err = svc.GetRecords(context.Background(), &models.QueryParams{
		PaginationFilterParams: models.PaginationFilterParams{Page: 1, Limit: 10},
		MinAmount:              100,
		MaxAmount:              2000,
		Search:                 "pay",
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || records[0].ID != "r2" {
		t.Errorf("unexpected filter result: total=%d", total)
	}
}

func TestGetGroupedRecords(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}, {Name: "Transport"}}); err != nil {
		t.Fatal(err)
	}

	_ = svc.CreateRecord(context.Background(), &models.Record{ID: "r1", Date: "2024-01-01", Description: "Lunch", Category: "Food", Amount: 20, Type: models.Expense})
	_ = svc.CreateRecord(context.Background(), &models.Record{ID: "r2", Date: "2024-01-02", Description: "Bus", Category: "Transport", Amount: 5, Type: models.Expense})
	_ = svc.CreateRecord(context.Background(), &models.Record{ID: "r3", Date: "2024-02-01", Description: "Dinner", Category: "Food", Amount: 30, Type: models.Expense})

	// Group by category
	groups, err := svc.GetGroupedRecords(context.Background(), &models.QueryParams{GroupBy: "category"})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	// Group by month
	groups, err = svc.GetGroupedRecords(context.Background(), &models.QueryParams{GroupBy: "month"})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 month groups, got %d", len(groups))
	}
}

func TestPatchRecord(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}, {Name: "Other"}}); err != nil {
		t.Fatal(err)
	}

	_ = svc.CreateRecord(context.Background(), &models.Record{ID: "r1", Date: "2024-01-01", Description: "Lunch", Category: "Food", Amount: 20, Type: models.Expense})

	newDesc := "Dinner"
	newAmount := 25.0
	newCat := "Other"
	err := svc.PatchRecord(context.Background(), "r1", &models.UpdateRecordRequest{
		Description: &newDesc,
		Amount:      &newAmount,
		Category:    &newCat,
	})
	if err != nil {
		t.Fatal(err)
	}

	rec, err := svc.GetRecord(context.Background(), "r1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Description != "Dinner" || rec.Amount != 25.0 || rec.Category != "other" {
		t.Errorf("patch didn't apply properly: %+v", rec)
	}
}

func TestExportRecords(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}}); err != nil {
		t.Fatal(err)
	}
	_ = svc.CreateRecord(context.Background(), &models.Record{ID: "r1", Date: "2024-01-01", Description: "Lunch", Category: "Food", Amount: 20, Type: models.Expense})

	recs, err := svc.ExportRecords(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 || recs[0].ID != "r1" {
		t.Errorf("export failed: %+v", recs)
	}
}

func TestImportCSVAndJSON(t *testing.T) {
	t.Run("JSON import", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		recs := []models.Record{
			{Date: "2024-01-01", Description: "Lunch", Category: "Food", Amount: 20, Type: models.Expense},
			{Date: "2024-01-02", Description: "Salary", Category: "Salary", Amount: 1000, Type: models.Income},
		}

		imported, err := svc.ImportJSON(context.Background(), recs)
		if err != nil {
			t.Fatal(err)
		}
		if imported != 2 {
			t.Errorf("expected 2 imported, got %d", imported)
		}

		var count int64
		db.Table("records").Count(&count)
		if count != 2 {
			t.Errorf("expected 2 records in DB, got %d", count)
		}
	})

	t.Run("CSV import", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		// Pre-create categories so mappings match
		if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Food"}}); err != nil {
			t.Fatal(err)
		}

		csvData := "date,description,amount,type,category\n2024-01-01,Lunch,20.0,expense,Food\n2024-01-02,Salary,1000.0,income,Salary\n"
		reader := strings.NewReader(csvData)

		imported, skipped, err := svc.ImportCSV(context.Background(), reader)
		if err != nil {
			t.Fatal(err)
		}
		if imported != 2 || skipped != 0 {
			t.Errorf("unexpected import result: imported=%d, skipped=%d", imported, skipped)
		}

		var count int64
		db.Table("records").Count(&count)
		if count != 2 {
			t.Errorf("expected 2 records in DB, got %d", count)
		}
	})
}

func TestCreateGoal(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Vacation"}}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}

	t.Run("success without category", func(t *testing.T) {
		goal := &models.Goal{Name: "Save for car", TargetAmount: 10000}
		if err := svc.CreateGoal(context.Background(), goal); err != nil {
			t.Fatalf("CreateGoal failed: %v", err)
		}
		if goal.ID == "" {
			t.Error("expected non-empty ID")
		}
		if goal.CurrentAmount != 0 {
			t.Errorf("expected currentAmount 0, got %f", goal.CurrentAmount)
		}
		if goal.Status != models.GoalActive {
			t.Errorf("expected active status, got %s", goal.Status)
		}
	})

	t.Run("success with category", func(t *testing.T) {
		goal := &models.Goal{Name: "Vacation fund", TargetAmount: 5000, Category: "vacation"}
		if err := svc.CreateGoal(context.Background(), goal); err != nil {
			t.Fatalf("CreateGoal failed: %v", err)
		}
		if goal.CategoryID == nil || *goal.CategoryID == "" {
			t.Error("expected categoryID to be set")
		}
	})

	t.Run("category not found", func(t *testing.T) {
		goal := &models.Goal{Name: "Bad goal", TargetAmount: 100, Category: "nonexistent"}
		err := svc.CreateGoal(context.Background(), goal)
		if err == nil {
			t.Fatal("expected error for nonexistent category")
		}
		if !isAppErrorType(err, "invalid_input") {
			t.Errorf("expected invalid_input error, got %v", err)
		}
	})
}

func TestGetGoals(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Travel"}}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}

	if err := svc.CreateGoal(context.Background(), &models.Goal{Name: "Europe trip", TargetAmount: 3000, Category: "travel"}); err != nil {
		t.Fatalf("CreateGoal setup failed: %v", err)
	}
	if err := svc.CreateGoal(context.Background(), &models.Goal{Name: "Emergency fund", TargetAmount: 5000}); err != nil {
		t.Fatalf("CreateGoal setup failed: %v", err)
	}

	goals, err := svc.GetGoals(context.Background())
	if err != nil {
		t.Fatalf("GetGoals failed: %v", err)
	}

	if len(goals) != 2 {
		t.Fatalf("expected 2 goals, got %d", len(goals))
	}

	for _, g := range goals {
		if g.ID == "" {
			t.Error("expected non-empty ID on goal")
		}
		if g.Name == "Europe trip" && g.Category != "travel" {
			t.Errorf("expected category 'travel', got %q", g.Category)
		}
	}
}

func TestGetGoal(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	goal := &models.Goal{Name: "Save", TargetAmount: 1000}
	if err := svc.CreateGoal(context.Background(), goal); err != nil {
		t.Fatalf("CreateGoal setup failed: %v", err)
	}

	t.Run("found", func(t *testing.T) {
		got, err := svc.GetGoal(context.Background(), goal.ID)
		if err != nil {
			t.Fatalf("GetGoal failed: %v", err)
		}
		if got.Name != "Save" {
			t.Errorf("expected name 'Save', got %q", got.Name)
		}
		if got.ID != goal.ID {
			t.Errorf("expected ID %q, got %q", goal.ID, got.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.GetGoal(context.Background(), "nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent goal")
		}
		if !isAppErrorType(err, "not_found") {
			t.Errorf("expected not_found error, got %v", err)
		}
	})
}

func TestUpdateGoal(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	if _, err := svc.CreateCategories(context.Background(), []models.Category{{Name: "Tech"}}); err != nil {
		t.Fatalf("CreateCategories setup failed: %v", err)
	}

	goal := &models.Goal{Name: "Laptop", TargetAmount: 2000}
	if err := svc.CreateGoal(context.Background(), goal); err != nil {
		t.Fatalf("CreateGoal setup failed: %v", err)
	}

	t.Run("update name", func(t *testing.T) {
		err := svc.UpdateGoal(context.Background(), goal.ID, &models.UpdateGoalRequest{Name: new("New Laptop")})
		if err != nil {
			t.Fatalf("UpdateGoal failed: %v", err)
		}
		got, _ := svc.GetGoal(context.Background(), goal.ID)
		if got.Name != "New Laptop" {
			t.Errorf("expected name 'New Laptop', got %q", got.Name)
		}
	})

	t.Run("update target amount", func(t *testing.T) {
		target := 3000.0
		err := svc.UpdateGoal(context.Background(), goal.ID, &models.UpdateGoalRequest{TargetAmount: &target})
		if err != nil {
			t.Fatalf("UpdateGoal failed: %v", err)
		}
		got, _ := svc.GetGoal(context.Background(), goal.ID)
		if got.TargetAmount != 3000 {
			t.Errorf("expected targetAmount 3000, got %f", got.TargetAmount)
		}
	})

	t.Run("update category", func(t *testing.T) {
		cat := "tech"
		err := svc.UpdateGoal(context.Background(), goal.ID, &models.UpdateGoalRequest{Category: &cat})
		if err != nil {
			t.Fatalf("UpdateGoal failed: %v", err)
		}
		got, _ := svc.GetGoal(context.Background(), goal.ID)
		if got.Category != "tech" {
			t.Errorf("expected category 'tech', got %q", got.Category)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.UpdateGoal(context.Background(), "nonexistent", &models.UpdateGoalRequest{Name: new("X")})
		if err == nil {
			t.Fatal("expected error for nonexistent goal")
		}
		if !isAppErrorType(err, "not_found") {
			t.Errorf("expected not_found error, got %v", err)
		}
	})
}

func TestDeleteGoal(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	goal := &models.Goal{Name: "To delete", TargetAmount: 100}
	if err := svc.CreateGoal(context.Background(), goal); err != nil {
		t.Fatalf("CreateGoal setup failed: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		if err := svc.DeleteGoal(context.Background(), goal.ID); err != nil {
			t.Fatalf("DeleteGoal failed: %v", err)
		}
		_, err := svc.GetGoal(context.Background(), goal.ID)
		if err == nil {
			t.Error("expected error after deletion")
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.DeleteGoal(context.Background(), "nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent goal")
		}
		if !isAppErrorType(err, "not_found") {
			t.Errorf("expected not_found error, got %v", err)
		}
	})
}

func TestAddGoalProgress(t *testing.T) {
	db := setupTestDB(t)
	svc := NewRecordService(db)

	goal := &models.Goal{Name: "Save", TargetAmount: 1000}
	if err := svc.CreateGoal(context.Background(), goal); err != nil {
		t.Fatalf("CreateGoal setup failed: %v", err)
	}

	t.Run("add progress", func(t *testing.T) {
		if err := svc.AddGoalProgress(context.Background(), goal.ID, 250); err != nil {
			t.Fatalf("AddGoalProgress failed: %v", err)
		}
		got, _ := svc.GetGoal(context.Background(), goal.ID)
		if got.CurrentAmount != 250 {
			t.Errorf("expected currentAmount 250, got %f", got.CurrentAmount)
		}
	})

	t.Run("add more progress", func(t *testing.T) {
		if err := svc.AddGoalProgress(context.Background(), goal.ID, 100); err != nil {
			t.Fatalf("AddGoalProgress failed: %v", err)
		}
		got, _ := svc.GetGoal(context.Background(), goal.ID)
		if got.CurrentAmount != 350 {
			t.Errorf("expected currentAmount 350, got %f", got.CurrentAmount)
		}
	})

	t.Run("not found", func(t *testing.T) {
		err := svc.AddGoalProgress(context.Background(), "nonexistent", 50)
		if err == nil {
			t.Fatal("expected error for nonexistent goal")
		}
		if !isAppErrorType(err, "not_found") {
			t.Errorf("expected not_found error, got %v", err)
		}
	})
}
