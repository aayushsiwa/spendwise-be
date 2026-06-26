package models

type Budget struct {
	ID         string  `gorm:"column:ID;primaryKey" json:"ID"`
	CategoryID string  `gorm:"column:categoryID;not null;constraint:OnDelete:CASCADE;uniqueIndex:idx_budget_unique,composite:categoryID" json:"categoryID"`
	Category   string  `gorm:"->;column:category" json:"category"`
	Month      int     `gorm:"column:month;not null;check:month >= 1 AND month <= 12;uniqueIndex:idx_budget_unique,composite:month" json:"month"`
	Year       int     `gorm:"column:year;not null;uniqueIndex:idx_budget_unique,composite:year" json:"year"`
	Amount     float64 `gorm:"column:amount;not null;check:amount > 0" json:"amount"`
}

func (Budget) TableName() string {
	return "budgets"
}

type BudgetProgress struct {
	Budget
	Spent      float64 `json:"spent"`
	Percentage float64 `json:"percentage"`
}
