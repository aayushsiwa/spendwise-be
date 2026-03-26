package models

type RecordType string

const (
	Income   RecordType = "income"
	Expense  RecordType = "expense"
	Transfer RecordType = "transfer"
)

type Record struct {
	ID          int        `json:"id"`
	Date        string     `json:"date"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Amount      float64    `json:"amount"`
	Type        RecordType `json:"type"`
	Note        string     `json:"note"`
	Balance     float64    `json:"balance"`
}

type RecordsResponse struct {
	Records []Record `json:"records"`
	PaginationMetadata
}
