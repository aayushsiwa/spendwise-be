package models

type Budget struct {
	ID         string  `gorm:"column:ID;primaryKey" json:"ID"`
	CategoryID string  `gorm:"column:categoryID;not null" json:"categoryID"`
	Category   string  `gorm:"->;column:category" json:"category"`
	Month      int     `gorm:"column:month;not null" json:"month"`
	Year       int     `gorm:"column:year;not null" json:"year"`
	Amount     float64 `gorm:"column:amount;not null" json:"amount"`
}

func (Budget) TableName() string {
	return "budgets"
}

type BudgetProgress struct {
	Budget
	Spent      float64 `json:"spent"`
	Percentage float64 `json:"percentage"`
}
