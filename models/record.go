package models

type Record struct {
	ID          int     `json:"id"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Note        string  `json:"note"`
	Balance     float64 `json:"balance"`
}
