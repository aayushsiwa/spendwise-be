package models

type CategoryDetail struct {
	ID         int     `json:"id"`
	CategoryID int     `json:"category_id"`
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
}

type Summary struct {
	Expenses     []CategoryDetail `json:"expenses"`
	Incomes      []CategoryDetail `json:"incomes"`
	Net          float64          `json:"net"`
	Opening      float64          `json:"opening"`
	Closing      float64          `json:"closing"`
	TotalExpense float64          `json:"total_expense"`
	TotalIncome  float64          `json:"total_income"`
}

type SummaryResponse struct {
	Summaries map[string]Summary `json:"summaries"`
	PaginationMetadata
}
