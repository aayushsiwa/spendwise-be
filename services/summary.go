package services

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"time"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/utils"
)

func (s *RecordService) UpdateSummary(ctx context.Context) (err error) {
	slog.InfoContext(ctx, "Updating summary...")

	tx, err := s.db.Begin()
	if err != nil {
		return errors.NewDatabase("Failed to begin transaction", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec("DELETE FROM summary"); err != nil {
		return errors.NewDatabase("Failed to clear summary", err)
	}
	if _, err = tx.Exec("DELETE FROM summary_details"); err != nil {
		return errors.NewDatabase("Failed to clear summary_details", err)
	}

	var minMonth sql.NullString
	if err = tx.QueryRow(`
		SELECT MIN(strftime('%Y-%m', date))
		FROM records
	`).Scan(&minMonth); err != nil {
		return errors.NewDatabase("Failed to get min month", err)
	}

	if !minMonth.Valid {
		slog.InfoContext(ctx, "No records found, summary will be empty")
		return tx.Commit()
	}

	maxMonth := time.Now().Format("2006-01")

	rows, err := tx.Query(`
		SELECT
			strftime('%Y-%m', date) AS month,
			SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END) AS "totalIncome",
			SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END) AS "totalExpense"
		FROM records
		GROUP BY month
		ORDER BY month ASC
	`)
	if err != nil {
		return errors.NewDatabase("Failed to aggregate records", err)
	}
	defer func() { _ = rows.Close() }()

	type monthData struct {
		income  float64
		expense float64
	}
	data := make(map[string]monthData)

	for rows.Next() {
		var m string
		var inc, exp float64
		if err = rows.Scan(&m, &inc, &exp); err != nil {
			return errors.NewDatabase("Failed to scan summary row", err)
		}
		data[m] = monthData{inc, exp}
	}
	if err = rows.Err(); err != nil {
		return errors.NewDatabase("Error iterating summary rows", err)
	}

	openingBalance := 0.0
	for m := minMonth.String; m <= maxMonth; m = utils.NextMonth(m) {
		d := data[m]
		net := d.income - d.expense
		closing := openingBalance + net

		_, err = tx.Exec(`
			INSERT INTO summary (month, "totalIncome", "totalExpense", "openingBalance", "netBalance", "closingBalance")
			VALUES (?, ?, ?, ?, ?, ?)
		`, m, d.income, d.expense, openingBalance, net, closing)
		if err != nil {
			return errors.NewDatabase("Failed to insert summary for month", err)
		}

		openingBalance = closing
	}

	_, err = tx.Exec(`
		INSERT INTO summary_details ("ID", month, type, "categoryID", "categoryName", amount)
		SELECT
			r.id,
			strftime('%Y-%m', r.date) AS month,
			r.type,
			COALESCE(c."ID", '') AS "categoryID",
			COALESCE(c.name, 'uncategorized') AS "categoryName",
			SUM(r.amount) AS amount
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c."ID"
		WHERE r.type IN ('income', 'expense', 'transfer')
		GROUP BY month, r.type, COALESCE(c."ID", ''), COALESCE(c.name, 'uncategorized')
	`)
	if err != nil {
		return errors.NewDatabase("Failed to insert summary details", err)
	}

	if err = tx.Commit(); err != nil {
		return errors.NewDatabase("Failed to commit summary transaction", err)
	}

	slog.InfoContext(ctx, "Summary updated successfully")
	return nil
}

func (s *RecordService) GetSummary(ctx context.Context, from, to, categoryFilter, typeFilter string) (*models.Summary, error) {
	var totalIncome, totalExpense float64
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM records WHERE date >= ? AND date <= ?
	`, from, to).Scan(&totalIncome, &totalExpense)
	if err != nil {
		return nil, errors.NewDatabase("Failed to compute totals", err)
	}

	var openingBalance float64
	err = s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(
			CASE WHEN type = 'income' THEN amount
			     WHEN type = 'expense' THEN -amount
			     ELSE 0 END
		), 0) FROM records WHERE date < ?
	`, from).Scan(&openingBalance)
	if err != nil {
		return nil, errors.NewDatabase("Failed to compute opening balance", err)
	}

	netInRange := totalIncome - totalExpense
	closingBalance := openingBalance + netInRange

	detailQuery := `
		SELECT COALESCE(c.id, ''), COALESCE(c.name, ''), r.type, SUM(r.amount)
		FROM records r
		LEFT JOIN categories c ON r."categoryID" = c.id
		WHERE r.date >= ? AND r.date <= ?
	`
	detailArgs := []any{from, to}

	var conditions []string
	if categoryFilter != "" {
		conditions = append(conditions, "LOWER(c.name) = ?")
		detailArgs = append(detailArgs, strings.ToLower(categoryFilter))
	}
	if typeFilter != "" {
		conditions = append(conditions, "r.type = ?")
		detailArgs = append(detailArgs, typeFilter)
	}
	if len(conditions) > 0 {
		detailQuery += " AND " + strings.Join(conditions, " AND ")
	}

	detailQuery += " GROUP BY c.id, r.type ORDER BY r.type, SUM(r.amount) DESC"

	rows, err := s.db.QueryContext(ctx, detailQuery, detailArgs...)
	if err != nil {
		return nil, errors.NewDatabase("Failed to fetch category breakdown", err)
	}
	defer func() { _ = rows.Close() }()

	incomes := make([]models.CategoryDetail, 0)
	expenses := make([]models.CategoryDetail, 0)
	for rows.Next() {
		var categoryID string
		var categoryName, recType string
		var amount float64
		if err := rows.Scan(&categoryID, &categoryName, &recType, &amount); err != nil {
			return nil, errors.NewDatabase("Failed to scan category detail", err)
		}
		cd := models.CategoryDetail{CategoryID: categoryID, Category: categoryName, Amount: amount}
		switch recType {
		case "income":
			incomes = append(incomes, cd)
		case "expense":
			expenses = append(expenses, cd)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabase("Error iterating category details", err)
	}

	return &models.Summary{
		Expenses:     expenses,
		Incomes:      incomes,
		Net:          netInRange,
		Opening:      openingBalance,
		Closing:      closingBalance,
		TotalExpense: totalExpense,
		TotalIncome:  totalIncome,
	}, nil
}
