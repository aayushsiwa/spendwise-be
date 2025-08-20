package models

type CategoryDetail struct {
	CategoryID int     `json:"category_id"`
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
}

type MonthlySummary struct {
	Expenses     []CategoryDetail `json:"expenses"`
	Incomes      []CategoryDetail `json:"incomes"`
	Net          float64          `json:"net"`
	Opening      float64          `json:"opening"`
	Closing      float64          `json:"closing"`
	TotalExpense float64          `json:"total_expenses"`
	TotalIncome  float64          `json:"total_income"`
}
