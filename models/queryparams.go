package models

// PaginationFilterParams is the structure of query params for pagination in get requests.
type PaginationFilterParams struct {
	Limit int `binding:"number,min=1" form:"limit,omitempty" json:"limit,omitempty"`
	Page  int `binding:"number,min=1" form:"page,omitempty"  json:"page,omitempty"`
}

type BaseQueryParams struct {
	From      string     `binding:"omitempty" form:"from" json:"from"`
	To        string     `binding:"omitempty" form:"to" json:"to"`
	Category  string     `binding:"omitempty" form:"category" json:"category"`
	Type      RecordType `binding:"omitempty,oneof=income expense transfer" form:"type" json:"type"`
	MinAmount float64    `binding:"omitempty" form:"minAmount" json:"minAmount"`
	MaxAmount float64    `binding:"omitempty" form:"maxAmount" json:"maxAmount"`
}
