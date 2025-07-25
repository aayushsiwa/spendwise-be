package utils

import (
	"fmt"
	"time"

	"aayushsiwa/expense-tracker/db"
)

func GenerateCustomID(date string) (string, error) {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}
	idPrefix := parsedDate.Format("020106") // ddmmyy

	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM records WHERE date = ?", date).Scan(&count)
	if err != nil {
		return "", err
	}

	customID := fmt.Sprintf("%s%d", idPrefix, count)
	return customID, nil
}
