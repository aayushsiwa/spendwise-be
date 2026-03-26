package handlers

import (
	"aayushsiwa/expense-tracker/utils"
	"database/sql"
	"encoding/csv"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type recordRow struct {
	date        interface{}
	description string
	categoryID  int
	amount      float64
	recordType  string
	note        string
	balance     float64
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, " ", "")
	return s
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func insertBatch(tx *sql.Tx, batch []recordRow) error {
	if len(batch) == 0 {
		return nil
	}

	query := `INSERT INTO records 
(date, description, category_id, amount, type, note, balance) 
VALUES `
	args := []interface{}{}
	values := make([]string, 0, len(batch))

	for _, r := range batch {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?)")
		args = append(args,
			r.date,
			r.description,
			r.categoryID,
			r.amount,
			r.recordType,
			r.note,
			0,
		)
	}

	query += strings.Join(values, ",")

	_, err := tx.Exec(query, args...)
	return err
}

func (h *Handler) ImportCSV(c *gin.Context) {
	ctx := c.Request.Context()
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file not provided"})
		return
	}

	if fileHeader.Size > 10<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format"})
		return
	}

	// ---------- FIELD MAPPING ----------
	fieldAliases := map[string][]string{
		"date":        {"date", "transactiondate", "txndate"},
		"description": {"description", "details", "narration"},
		"category":    {"category", "type"},
		"amount":      {"amount", "value", "amt"},
		"type":        {"type", "transactiontype"},
		"note":        {"note", "remarks", "comment"},
	}

	headerIndex := make(map[string]int)
	for i, h := range headers {
		headerIndex[normalize(h)] = i
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing date column"})
		return
	}
	if _, ok := fieldIndex["amount"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing amount column"})
		return
	}

	getValue := func(record []string, field string) string {
		idx, ok := fieldIndex[field]
		if !ok || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[idx])
	}

	// ---------- START TRANSACTION ----------
	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// ---------- PRELOAD CATEGORIES ----------
	categoryMap := make(map[string]int)

	rows, err := tx.Query(`SELECT id, name FROM categories`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name string
			if err := rows.Scan(&id, &name); err == nil {
				categoryMap[strings.ToLower(name)] = id
			}
		}
	}

	// ---------- PROCESS CSV ----------
	var importedCount, skippedCount int
	var batch []recordRow
	batchSize := 100

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading CSV"})
			return
		}

		dateStr := getValue(record, "date")
		description := getValue(record, "description")
		category := strings.ToLower(getValue(record, "category"))
		amountStr := getValue(record, "amount")
		recordType := getValue(record, "type")
		note := getValue(record, "note")

		if dateStr == "" || amountStr == "" {
			skippedCount++
			continue
		}

		date, err := utils.ParseDate(dateStr)
		if err != nil {
			skippedCount++
			continue
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			skippedCount++
			continue
		}

		if recordType == "" {
			if amount < 0 {
				recordType = "expense"
			} else {
				recordType = "income"
			}
		}

		amount = abs(amount)

		if category == "" {
			category = "uncategorized"
		}

		categoryID, ok := categoryMap[category]
		if !ok {
			res, err := tx.Exec(`INSERT INTO categories (name) VALUES (?)`, category)
			if err != nil {
				skippedCount++
				continue
			}
			lastID, _ := res.LastInsertId()
			categoryID = int(lastID)
			categoryMap[category] = categoryID
		}

		batch = append(batch, recordRow{
			date:        date,
			description: description,
			categoryID:  categoryID,
			amount:      amount,
			recordType:  recordType,
			note:        note,
		})

		if len(batch) >= batchSize {
			if err := insertBatch(tx, batch); err != nil {
				slog.ErrorContext(ctx, "Failed to insert batch", slog.Any("error", err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Batch insert failed"})
				return
			}
			importedCount += len(batch)
			batch = nil
		}
	}

	if len(batch) > 0 {
		if err := insertBatch(tx, batch); err != nil {
			slog.ErrorContext(ctx, "Failed to insert batch", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Batch insert failed"})
			return
		}
		importedCount += len(batch)
	}

	// ---------- RECALCULATE BALANCES ----------
	if err := h.recalculateBalances(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to recalculate balances", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Balance recalculation failed"})
		return
	}

	// ---------- COMMIT ----------
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Commit failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "CSV import completed successfully",
		"recordsImported": importedCount,
		"skippedCount":    skippedCount,
	})
}
