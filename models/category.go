package models

type Category struct {
	ID    string `gorm:"primaryKey;column:ID" json:"ID"`
	Name  string `gorm:"column:name;uniqueIndex;not null" json:"name"`
	Icon  string `gorm:"column:icon" json:"icon"`
	Color string `gorm:"column:color" json:"color"`
}

func (Category) TableName() string {
	return "categories"
}
