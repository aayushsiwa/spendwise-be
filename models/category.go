package models

type Category struct {
	ID    int    `json:"ID"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}
