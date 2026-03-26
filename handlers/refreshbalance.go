package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RecalculateBalances(c *gin.Context) {
	ctx := c.Request.Context()
	tx, err := h.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			slog.ErrorContext(ctx, "Failed to rollback transaction", err)
		}
	}(tx)

	if err = h.recalculateBalances(ctx, tx); err != nil {
		appErr := fmt.Errorf("failed to recalculate balances: %w", err)
		slog.ErrorContext(ctx, "Error in RecalculateBalances: %v", appErr)
		c.JSON(500, gin.H{"error": "Internal Server Error"})
		return
	}

	if err := tx.Commit(); err != nil {
		slog.ErrorContext(ctx, "Failed to commit transaction", err)
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
			slog.ErrorContext(ctx, "Error closing rows: %v", err)
		}
	}(rows)

	var (
		runningBalance float64
		batchSize      = 100
		ids            []int
		balances       []float64
	)

	updateBatch := func() error {
		if len(ids) == 0 {
			return nil
		}

		query := "UPDATE records SET balance = CASE id "
		var args []interface{}

		for i := range ids {
			query += "WHEN ? THEN ? "
			args = append(args, ids[i], balances[i])
		}

		query += "END WHERE id IN ("
		for i := range ids {
			if i > 0 {
				query += ","
			}
			query += "?"
			args = append(args, ids[i])
		}
		query += ")"

		_, err := tx.Exec(query, args...)
		return err
	}

	for rows.Next() {
		var id int
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
