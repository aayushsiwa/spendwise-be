package models

type UpdateRecordRequest struct {
	Date        *string  `json:"date,omitempty"`
	Description *string  `json:"description,omitempty"`
	Category    *string  `json:"category,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	Type        *string  `json:"type,omitempty"`
	Note        *string  `json:"note,omitempty"`
}
