package services

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "modernc.org/sqlite"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

	"github.com/lithammer/shortuuid/v4"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	t.Cleanup(func() { _ = db.Close() })

	schema, err := os.ReadFile("../sql/init.sql")
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("failed to execute schema: %v", err)
	}
	return db
}

func isAppErrorType(err error, typ string) bool {
	var appErr *apperrors.AppError
	return errors.As(err, &appErr) && appErr.Type == typ
}

func TestBuildWhereClause(t *testing.T) {
	t.Run("type filter", func(t *testing.T) {
		where, args := BuildWhereClause(&models.QueryParams{Type: "income"})
		if where != " WHERE r.type = ?" {
			t.Errorf("got %q", where)
		}
		if len(args) != 1 || args[0] != models.RecordType("income") {
			t.Errorf("unexpected args: %v", args)
		}
	})
	t.Run("no filters", func(t *testing.T) {
		where, args := BuildWhereClause(&models.QueryParams{})
		if where != "" {
			t.Errorf("got %q", where)
		}
		if len(args) != 0 {
			t.Errorf("unexpected args: %v", args)
		}
	})
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
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM summary").Scan(&count)
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
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM summary").Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count == 0 {
			t.Errorf("expected summary rows, got 0")
		}

		// Verify summary details rows
		err = db.QueryRow("SELECT COUNT(*) FROM summary_details").Scan(&count)
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

		_, _ = db.Exec("DROP TABLE summary")

		err := svc.UpdateSummary(context.Background())
		if err == nil {
			t.Fatal("expected error when summary table is missing")
		}
		if !isAppErrorType(err, "database") {
			t.Errorf("expected database error for summary clear failure, got %T: %v", err, err)
		}
	})

	t.Run("failed to clear summary_details", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		_, _ = db.Exec("DROP TABLE summary_details")

		err := svc.UpdateSummary(context.Background())
		if err == nil {
			t.Fatal("expected error when summary_details table is missing")
		}
		if !isAppErrorType(err, "database") {
			t.Errorf("expected database error for summary_details clear failure, got %T: %v", err, err)
		}
	})

	t.Run("failed to get min month", func(t *testing.T) {
		db := setupTestDB(t)
		svc := NewRecordService(db)

		_, _ = db.Exec("DROP TABLE records")

		err := svc.UpdateSummary(context.Background())
		if err == nil {
			t.Fatal("expected error when records table is missing")
		}
		if !isAppErrorType(err, "database") {
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
		if _, err := db.Exec("DROP TABLE summary"); err != nil {
			t.Fatalf("failed to drop summary table: %v", err)
		}
		if _, err := db.Exec("CREATE TABLE summary (month TEXT PRIMARY KEY CHECK (month = 'invalid'))"); err != nil {
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
		if _, err := db.Exec("DROP TABLE summary_details"); err != nil {
			t.Fatalf("failed to drop summary_details table: %v", err)
		}
		if _, err := db.Exec("CREATE TABLE summary_details (month TEXT, type TEXT, categoryID TEXT, categoryName TEXT, amount REAL, PRIMARY KEY (month, type, categoryID), CHECK (amount < 0))"); err != nil {
			t.Fatalf("failed to create mock summary_details table: %v", err)
		}

		err := svc.UpdateSummary(context.Background())
		if err == nil {
			t.Error("expected error due to CHECK constraint failure on summary_details table")
		}
	})
}
