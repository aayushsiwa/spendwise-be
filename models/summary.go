package models

type CategoryDetail struct {
	ID         string  `json:"ID"`
	CategoryID string  `json:"categoryID"`
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
}

type Summary struct {
	Expenses     []CategoryDetail `json:"expenses"`
	Incomes      []CategoryDetail `json:"incomes"`
	Net          float64          `json:"net"`
	Opening      float64          `json:"opening"`
	Closing      float64          `json:"closing"`
	TotalExpense float64          `json:"totalExpense"`
	TotalIncome  float64          `json:"totalIncome"`
}

type SummaryResponse struct {
	Summaries map[string]Summary `json:"summaries"`
	PaginationMetadata
}
