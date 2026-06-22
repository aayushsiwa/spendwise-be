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

type SummaryDB struct {
	Month          string  `gorm:"primaryKey;column:month" json:"month"`
	TotalIncome    float64 `gorm:"column:totalIncome;not null;default:0" json:"totalIncome"`
	TotalExpense   float64 `gorm:"column:totalExpense;not null;default:0" json:"totalExpense"`
	OpeningBalance float64 `gorm:"column:openingBalance;not null;default:0" json:"openingBalance"`
	NetBalance     float64 `gorm:"column:netBalance;not null;default:0" json:"netBalance"`
	ClosingBalance float64 `gorm:"column:closingBalance;not null;default:0" json:"closingBalance"`
}

func (SummaryDB) TableName() string {
	return "summary"
}

type SummaryDetailDB struct {
	Month        string  `gorm:"primaryKey;column:month" json:"month"`
	Type         string  `gorm:"primaryKey;column:type" json:"type"`
	CategoryID   string  `gorm:"primaryKey;column:categoryID" json:"categoryID"`
	CategoryName string  `gorm:"column:categoryName;not null" json:"categoryName"`
	Amount       float64 `gorm:"column:amount;not null" json:"amount"`
}

func (SummaryDetailDB) TableName() string {
	return "summary_details"
}
