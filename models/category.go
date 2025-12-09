package models

type Category struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}
