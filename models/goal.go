package models

type GoalStatus string

const (
	GoalActive    GoalStatus = "active"
	GoalAchieved  GoalStatus = "achieved"
	GoalAbandoned GoalStatus = "abandoned"
)

type Goal struct {
	ID                  string     `gorm:"primaryKey;column:ID" json:"ID"`
	Name                string     `gorm:"column:name;not null" json:"name"`
	TargetAmount        float64    `gorm:"column:targetAmount;not null" json:"targetAmount"`
	CurrentAmount       float64    `gorm:"column:currentAmount;not null;default:0" json:"currentAmount"`
	TargetDate          string     `gorm:"column:targetDate" json:"targetDate,omitempty"`
	Category            string     `gorm:"->;column:category" json:"category,omitempty"`
	CategoryID          *string    `gorm:"column:categoryID" json:"categoryID,omitempty"`
	Status              GoalStatus `gorm:"column:status;not null;default:active" json:"status"`
	Description         string     `gorm:"column:description" json:"description,omitempty"`
	MonthlyContribution float64    `gorm:"column:monthlyContribution;default:0" json:"monthlyContribution,omitempty"`
}

func (Goal) TableName() string {
	return "goals"
}

type UpdateGoalRequest struct {
	Name                *string  `json:"name,omitempty"`
	TargetAmount        *float64 `json:"targetAmount,omitempty"`
	CurrentAmount       *float64 `json:"currentAmount,omitempty"`
	TargetDate          *string  `json:"targetDate,omitempty"`
	Category            *string  `json:"category,omitempty"`
	Status              *string  `json:"status,omitempty"`
	Description         *string  `json:"description,omitempty"`
	MonthlyContribution *float64 `json:"monthlyContribution,omitempty"`
}

type AddProgressRequest struct {
	Amount float64 `json:"amount"`
}
