package models

type Category struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type Record struct {
	ID          int     `json:"id"`
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Notes       string  `json:"notes"`
}
