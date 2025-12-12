package models

// PaginationFilterParams is the structure of query params for pagination in get requests.
type PaginationFilterParams struct {
	Limit int `binding:"required,number,min=1" form:"limit,omitempty" json:"limit,omitempty"`
	Page  int `binding:"required,number,min=1" form:"page,omitempty"  json:"page,omitempty"`
}

type TimeFrame string

const (
	Year    TimeFrame = "year"
	Quarter TimeFrame = "quarter"
	Month   TimeFrame = "month"
)

// TimeFrameParams is the structure of query params for timeframe-based requests.
type TimeFrameParams struct {
	TimeFrame string `binding:"omitempty,oneof=year quarter month" form:"timeframe" json:"timeframe,omitempty"`
	Year      string `binding:"omitempty,number" form:"year" json:"year,omitempty"`
	Quarter   string `binding:"omitempty,number,min=1,max=4" form:"quarter" json:"quarter,omitempty"`
	Month     string `binding:"omitempty,number,min=1,max=12" form:"month" json:"month,omitempty"`
}

// QueryParams is the structure of query params for filtering records.
// When adding new fields here, ensure to update utils/buildwhereclause.go accordingly.
type QueryParams struct {
	PaginationFilterParams
	TimeFrameParams
	From      string     `binding:"omitempty" form:"from" json:"from"`
	To        string     `binding:"omitempty" form:"to" json:"to"`
	Category  string     `binding:"omitempty" form:"category" json:"category"`
	Type      RecordType `binding:"omitempty,oneof=income expense transfer" form:"type" json:"type"`
	MinAmount float64    `binding:"omitempty" form:"minAmount" json:"minAmount"`
	MaxAmount float64    `binding:"omitempty" form:"maxAmount" json:"maxAmount"`
	GroupBy   string     `binding:"omitempty,oneof=category month" form:"groupBy" json:"groupBy"`
	Search    string     `binding:"omitempty" form:"search" json:"search"`
}

type PaginationMetadata struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	TotalPages int  `json:"total_pages,omitempty"`
	TotalCount int  `json:"total_count,omitempty"`
	HasPrev    bool `json:"has_prev,omitempty"`
	HasNext    bool `json:"has_next,omitempty"`
}
