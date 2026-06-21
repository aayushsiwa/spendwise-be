package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RecalculateBalances(c *gin.Context) {
	ctx := c.Request.Context()
	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = h.recalculateBalances(ctx, tx); err != nil {
		slog.ErrorContext(ctx, "Failed to recalculate balances", "error", err)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}

	if err = tx.Commit(); err != nil {
		slog.ErrorContext(ctx, "Failed to commit transaction", "error", err)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(200, gin.H{"status": "Balances recalculated successfully"})
}

func (h *Handler) recalculateBalances(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.Query(`
		SELECT id, date, amount, type
		FROM records
		ORDER BY date ASC, id ASC
	`)
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			slog.ErrorContext(ctx, "Error closing rows", "err", err)
		}
	}(rows)

	var (
		runningBalance float64
		batchSize      = 100
		ids            []string
		balances       []float64
	)

	updateBatch := func() error {
		if len(ids) == 0 {
			return nil
		}

		var query strings.Builder
		query.WriteString("UPDATE records SET balance = CASE id ")
		var args []any

		for i := range ids {
			query.WriteString("WHEN ? THEN ? ")
			args = append(args, ids[i], balances[i])
		}

		query.WriteString("END WHERE id IN (")
		for i := range ids {
			if i > 0 {
				query.WriteString(",")
			}
			query.WriteString("?")
			args = append(args, ids[i])
		}
		query.WriteString(")")

		_, err := tx.Exec(query.String(), args...)
		return err
	}

	for rows.Next() {
		var id string
		var date string
		var amount float64
		var recordType string

		if err := rows.Scan(&id, &date, &amount, &recordType); err != nil {
			return err
		}

		if recordType == "income" {
			runningBalance += amount
		} else {
			runningBalance -= amount
		}

		ids = append(ids, id)
		balances = append(balances, runningBalance)

		if len(ids) >= batchSize {
			if err := updateBatch(); err != nil {
				return err
			}
			ids = nil
			balances = nil
		}
	}

	// flush remaining
	if err := updateBatch(); err != nil {
		return err
	}

	return nil
}
