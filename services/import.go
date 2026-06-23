package services

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"log/slog"
	"strconv"
	"strings"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/utils"

	"github.com/lithammer/shortuuid/v4"
	"gorm.io/gorm"
)

func (s *RecordService) ImportCSV(ctx context.Context, src io.Reader) (imported, skipped int, err error) {
	reader := csv.NewReader(src)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return 0, 0, errors.Join(ErrImportValidation, err)
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

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, 0, tx.Error
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var categories []models.Category
	if err = tx.Find(&categories).Error; err != nil {
		return 0, 0, apperrors.NewDatabase("Failed to fetch categories", err)
	}
	categoryMap := make(map[string]string)
	for _, c := range categories {
		categoryMap[strings.ToLower(strings.TrimSpace(c.Name))] = c.ID
	}

	var batch []models.Record
	batchSize := 100

	for {
		record, readErr := reader.Read()
		if readErr == io.EOF {
			break
		} else if readErr != nil {
			err = errors.Join(ErrImportValidation, readErr)
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

		var catID *string
		if category != "" {
			if id, ok := categoryMap[category]; ok {
				catID = &id
			}
		}

		batch = append(batch, models.Record{
			ID:          shortuuid.New(),
			Date:        date,
			Description: description,
			CategoryID:  catID,
			Amount:      amount,
			Type:        models.RecordType(recordType),
			Note:        note,
			Balance:     0,
		})

		if len(batch) >= batchSize {
			if err = tx.Create(&batch).Error; err != nil {
				slog.ErrorContext(ctx, "Failed to insert batch", slog.Any("error", err))
				return
			}
			imported += len(batch)
			batch = nil
		}
	}

	if len(batch) > 0 {
		if err = tx.Create(&batch).Error; err != nil {
			slog.ErrorContext(ctx, "Failed to insert batch", slog.Any("error", err))
			return
		}
		imported += len(batch)
	}

	if err = recalculateBalances(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to recalculate balances", slog.Any("error", err))
		return
	}

	if err = s.updateSummaryTx(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to update summary after CSV import", "error", err)
		return
	}

	if err = tx.Commit().Error; err != nil {
		return
	}

	return
}

func (s *RecordService) ImportJSON(ctx context.Context, records []models.Record) (imported int, err error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer func() {
		if err != nil {
			tx.Rollback()
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

		var cat models.Category
		err = tx.Where("LOWER(name) = ?", category).First(&cat).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				cat = models.Category{
					ID:   shortuuid.New(),
					Name: category,
				}
				if err = tx.Create(&cat).Error; err != nil {
					continue
				}
			} else {
				continue
			}
		}

		rec.ID = shortuuid.New()
		rec.Date = date
		rec.CategoryID = &cat.ID
		rec.Balance = 0

		if err = tx.Create(&rec).Error; err != nil {
			continue
		}

		imported++
	}

	if imported == 0 {
		tx.Rollback()
		return imported, nil
	}

	if err = recalculateBalances(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to recalculate balances", "error", err)
		return
	}

	if err = s.updateSummaryTx(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to update summary after JSON import", "error", err)
		return
	}

	if err = tx.Commit().Error; err != nil {
		return
	}

	return
}
