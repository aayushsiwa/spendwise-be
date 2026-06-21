package services

import (
	"context"
	"database/sql"
	"encoding/csv"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/utils"

	"github.com/lithammer/shortuuid/v4"
)

func insertBatch(tx *sql.Tx, batch []recordRow) error {
	if len(batch) == 0 {
		return nil
	}

	query := `INSERT INTO records
(id, date, description, "categoryID", amount, type, note, balance)
VALUES `
	args := []any{}
	values := make([]string, 0, len(batch))

	for _, r := range batch {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?, 0)")
		args = append(args,
			r.id,
			r.date,
			r.description,
			r.categoryID,
			r.amount,
			r.recordType,
			r.note,
		)
	}

	query += strings.Join(values, ",")

	_, err := tx.Exec(query, args...)
	return err
}

func (s *RecordService) ImportCSV(ctx context.Context, src io.Reader) (imported, skipped int, err error) {
	reader := csv.NewReader(src)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return 0, 0, err
	}

	fieldAliases := map[string][]string{
		"date":        {"date", "transactiondate", "txndate", "valuedate", "postingdate"},
		"description": {"description", "details", "narration", "transactiondescription", "transactiondetails", "particulars"},
		"amount":      {"amount", "value", "amt", "transactionamount", "txnamount"},
		"type":        {"type", "transactiontype", "drcr", "dr/cr", "debitcredit", "debit/credit"},
		"note":        {"note", "remarks", "comment", "reference", "chq/refno", "chqrefno"},
		"category":    {"category", "subcategory", "sub-category"},
	}

	headerIndex := make(map[string]int)
	for i, h := range headers {
		headerIndex[utils.Normalize(h)] = i
	}

	fieldIndex := make(map[string]int)
	for field, aliases := range fieldAliases {
		for _, alias := range aliases {
			if idx, ok := headerIndex[alias]; ok {
				fieldIndex[field] = idx
				break
			}
		}
	}

	if _, ok := fieldIndex["date"]; !ok {
		return 0, 0, errMissingColumn("date")
	}
	if _, ok := fieldIndex["amount"]; !ok {
		return 0, 0, errMissingColumn("amount")
	}

	getValue := func(record []string, field string) string {
		idx, ok := fieldIndex[field]
		if !ok || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[idx])
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	categoryMap := make(map[string]string)
	rows, err := tx.QueryContext(ctx, `SELECT "ID", name FROM categories`)
	if err == nil {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var id string
			var name string
			if err := rows.Scan(&id, &name); err == nil {
				categoryMap[strings.ToLower(strings.TrimSpace(name))] = id
			}
		}
	}

	var batch []recordRow
	batchSize := 100

	for {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		} else if readErr != nil {
			err = readErr
			return
		}

		dateStr := getValue(record, "date")
		description := getValue(record, "description")
		amountStr := getValue(record, "amount")
		recordType := getValue(record, "type")
		note := getValue(record, "note")
		category := strings.ToLower(strings.TrimSpace(getValue(record, "category")))

		if dateStr == "" || amountStr == "" {
			skipped++
			continue
		}

		date, parseErr := utils.ParseDate(dateStr)
		if parseErr != nil {
			skipped++
			continue
		}

		amount, parseErr := strconv.ParseFloat(amountStr, 64)
		if parseErr != nil {
			skipped++
			continue
		}

		if recordType != "" {
			recordType = utils.InferRecordType(recordType)
		}
		if recordType == "" {
			if amount < 0 {
				recordType = "expense"
			} else {
				recordType = "income"
			}
		}

		amount = utils.Abs(amount)

		var catID any
		if category != "" {
			if id, ok := categoryMap[category]; ok {
				catID = id
			}
		}

		batch = append(batch, recordRow{
			id:          shortuuid.New(),
			date:        date,
			description: description,
			categoryID:  catID,
			amount:      amount,
			recordType:  recordType,
			note:        note,
		})

		if len(batch) >= batchSize {
			if err = insertBatch(tx, batch); err != nil {
				slog.ErrorContext(ctx, "Failed to insert batch", slog.Any("error", err))
				// err is set, defer will rollback
				return
			}
			imported += len(batch)
			batch = nil
		}
	}

	if len(batch) > 0 {
		if err = insertBatch(tx, batch); err != nil {
			slog.ErrorContext(ctx, "Failed to insert batch", slog.Any("error", err))
			return
		}
		imported += len(batch)
	}

	if err = recalculateBalances(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to recalculate balances", slog.Any("error", err))
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	return
}

func (s *RecordService) ImportJSON(ctx context.Context, records []models.Record) (imported int, err error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.ErrorContext(ctx, "Failed to rollback transaction", "error", rbErr)
			}
		}
	}()

	for _, rec := range records {
		if rec.Date == "" || rec.Description == "" || rec.Category == "" || rec.Type == "" {
			continue
		}

		category := strings.ToLower(strings.TrimSpace(rec.Category))
		dateStr := strings.TrimSpace(rec.Date)

		date, parseErr := utils.ParseDate(dateStr)
		if parseErr != nil {
			slog.ErrorContext(ctx, "Failed to parse date", "date", dateStr, "error", parseErr)
			continue
		}

		var categoryID string
		err = tx.QueryRowContext(ctx, `SELECT "ID" FROM categories WHERE name = ?`, category).Scan(&categoryID)
		if err != nil {
			catID := shortuuid.New()
			_, err = tx.ExecContext(ctx, `INSERT INTO categories ("ID", name) VALUES (?, ?)`, catID, category)
			if err != nil {
				continue
			}
			categoryID = catID
		}

		var currentBalance float64
		err = tx.QueryRowContext(ctx, "SELECT COALESCE(balance, 0) FROM records ORDER BY date DESC, id DESC LIMIT 1").Scan(&currentBalance)
		if err != nil {
			currentBalance = 0
		}

		switch rec.Type {
		case "income":
			currentBalance += rec.Amount
		case "expense":
			currentBalance -= rec.Amount
		}

		customID := shortuuid.New()

		_, err = tx.ExecContext(ctx, `INSERT INTO records (id, date, description, "categoryID", amount, type, note, balance) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			customID, date, rec.Description, categoryID, rec.Amount, rec.Type, rec.Note, currentBalance)
		if err != nil {
			continue
		}

		imported++
	}

	if err = recalculateBalances(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to recalculate balances", "error", err)
	}

	if err = tx.Commit(); err != nil {
		return
	}

	if err = s.UpdateSummary(ctx); err != nil {
		slog.WarnContext(ctx, "Failed to update summary after JSON import", "error", err)
	}

	return
}
