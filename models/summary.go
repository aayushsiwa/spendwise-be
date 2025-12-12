package models

type CategoryDetail struct {
	ID         int     `json:"id"`
	CategoryID int     `json:"category_id"`
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
}

type Summary struct {
	Expenses     []CategoryDetail `json:"expenses,omitempty"`
	Incomes      []CategoryDetail `json:"incomes,omitempty"`
	Net          float64          `json:"net,omitempty"`
	Opening      float64          `json:"opening,omitempty"`
	Closing      float64          `json:"closing,omitempty"`
	TotalExpense float64          `json:"total_expense,omitempty"`
	TotalIncome  float64          `json:"total_income,omitempty"`
}

type SummaryResponse struct {
	Summaries map[string]Summary `json:"summaries"`
	PaginationMetadata
}
