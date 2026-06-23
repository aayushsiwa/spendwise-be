package models

import "testing"

func TestBudgetTableName(t *testing.T) {
	b := Budget{}
	if got := b.TableName(); got != "budgets" {
		t.Errorf("Budget.TableName() = %q, want %q", got, "budgets")
	}
}

func TestBudgetProgressEmbedsBudget(t *testing.T) {
	budget := Budget{
		ID:         "b1",
		CategoryID: "cat1",
		Month:      6,
		Year:       2026,
		Amount:     500.0,
	}
	progress := BudgetProgress{
		Budget:     budget,
		Spent:      300.0,
		Percentage: 60.0,
	}

	if progress.ID != "b1" {
		t.Errorf("BudgetProgress.ID = %q, want %q", progress.ID, "b1")
	}
	if progress.Amount != 500.0 {
		t.Errorf("BudgetProgress.Amount = %f, want %f", progress.Amount, 500.0)
	}
	if progress.Spent != 300.0 {
		t.Errorf("BudgetProgress.Spent = %f, want %f", progress.Spent, 300.0)
	}
	if progress.Percentage != 60.0 {
		t.Errorf("BudgetProgress.Percentage = %f, want %f", progress.Percentage, 60.0)
	}
}

func TestBudgetZeroValue(t *testing.T) {
	var b Budget
	if b.ID != "" {
		t.Errorf("zero-value Budget.ID = %q, want empty", b.ID)
	}
	if b.Amount != 0 {
		t.Errorf("zero-value Budget.Amount = %f, want 0", b.Amount)
	}
}