package handlers

import (
	"aayushsiwa/expense-tracker/models"
	"database/sql"
	"fmt"
	"log"
	"time"
)

func (h *Handler) RecalculateBalances() error {
	// Step 1: Fetch all records ordered by date
	rows, err := h.DB.Query(`
		SELECT id, date, amount, type 
		FROM records 
		ORDER BY date ASC
	`)
	if err != nil {
		return fmt.Errorf("failed to fetch records: %w", err)
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var r models.Record
		if err := rows.Scan(&r.ID, &r.Date, &r.Amount, &r.Type); err != nil {
			return err
		}
		records = append(records, r)
	}

	if len(records) == 0 {
		return nil
	}

	// Step 2: Iterate and update balances
	var runningBalance float64
	var currentMonth string

	for _, rec := range records {
		// Parse date to time.Time
		parsedDate, err := time.Parse("2006-01-02", rec.Date[:10])
		if err != nil {
			log.Printf("Skipping record %d: invalid date %s", rec.ID, rec.Date)
			continue
		}

		monthKey := parsedDate.Format("2006-01")

		// When month changes, initialize runningBalance from previous month's closing_balance
		if monthKey != currentMonth {
			currentMonth = monthKey

			// Find previous month key
			prevMonth := parsedDate.AddDate(0, -1, 0).Format("2006-01")

			// Try to get previous month closing_balance
			var prevClosing float64
			err := h.DB.QueryRow(`
				SELECT closing_balance 
				FROM summary 
				WHERE month = ?
			`, prevMonth).Scan(&prevClosing)

			if err == sql.ErrNoRows {
				runningBalance = 0 // No data → start from 0
			} else if err != nil {
				return fmt.Errorf("failed to get summary for %s: %w", prevMonth, err)
			} else {
				runningBalance = prevClosing
			}

			log.Printf("Initialized month %s with starting balance %.2f", monthKey, runningBalance)
		}

		// Update running balance based on record type
		if rec.Type == "income" {
			runningBalance += rec.Amount
		} else if rec.Type == "expense" {
			runningBalance -= rec.Amount
		}

		// Update the record's balance in DB
		_, err = h.DB.Exec(`
			UPDATE records SET balance = ? WHERE id = ?
		`, runningBalance, rec.ID)
		if err != nil {
			return fmt.Errorf("failed to update balance for record %d: %w", rec.ID, err)
		}
	}

	return nil
}
