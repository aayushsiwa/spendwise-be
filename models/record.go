package models

type RecordType string

const (
	Income   RecordType = "income"
	Expense  RecordType = "expense"
	Transfer RecordType = "transfer"
)

type Record struct {
	ID          string     `gorm:"primaryKey;column:ID" json:"ID"`
	Date        string     `gorm:"column:date;not null" json:"date"`
	Description string     `gorm:"column:description;not null" json:"description"`
	CategoryID  *string    `gorm:"column:categoryID" json:"categoryID,omitempty"`
	Category    string     `gorm:"-" json:"category"`
	Amount      float64    `gorm:"column:amount;not null" json:"amount"`
	Type        RecordType `gorm:"column:type;not null" json:"type"`
	Note        string     `gorm:"column:note" json:"note"`
	Balance     float64    `gorm:"column:balance;not null" json:"balance"`
}

func (Record) TableName() string {
	return "records"
}

type RecordsResponse struct {
	Records []Record `json:"records"`
	PaginationMetadata
}

type GroupedRecord struct {
	Group string  `json:"group"`
	Total float64 `json:"total"`
	Count int     `json:"count"`
}

type GroupedResponse struct {
	Groups []GroupedRecord `json:"groups"`
}
