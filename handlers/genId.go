package handlers

import (
	"database/sql"
	"fmt"
	"time"
)

func (h *Handler) GenerateCustomID(date string) (string, error) {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}
	idPrefix := parsedDate.Format("060102") // yymmdd

	var customID string
	var count int

	// Start with the number of records on that date
	err = h.DB.QueryRow("SELECT COUNT(*) FROM records WHERE date = ?", date).Scan(&count)
	if err != nil {
		return "", err
	}

	for {
		customID = fmt.Sprintf("%s%d", idPrefix, count)

		var exists string
		err := h.DB.QueryRow("SELECT id FROM records WHERE id = ?", customID).Scan(&exists)
		if err == sql.ErrNoRows {
			// ID is unique
			break
		} else if err != nil {
			// Some other error occurred
			return "", err
		}
		// ID exists, try next count
		count++
	}

	return customID, nil
}
