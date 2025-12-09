package handlers

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/secure"
	"aayushsiwa/expense-tracker/utils"
	"aayushsiwa/expense-tracker/validation"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetRecords(c *gin.Context) {
	// Query params
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	category := c.Query("category")
	recordType := c.Query("type")
	desc := c.Query("description")
	minAmountStr := c.Query("min_amount")
	maxAmountStr := c.Query("max_amount")
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	slog.Info("GetRecords called with filters",
		"start_date", startDate,
		"end_date", endDate,
		"category", category,
		"type", recordType,
		"description", desc,
		"min_amount", minAmountStr,
		"max_amount", maxAmountStr,
		"page", pageStr,
		"limit", limitStr,
	)

	// Validation
	validator := validation.NewValidator()
	var validationErrs errors.ValidationErrors

	// Parse pagination parameters
	page := 1
	limit := 50 // default limit

	if pageStr != "" {
		pageNum, err := strconv.Atoi(pageStr)
		if err != nil || pageNum < 1 {
			validationErrs = append(validationErrs, errors.NewValidationError("page", "Page must be a positive integer", pageStr))
		} else {
			page = pageNum
		}
	}

	if limitStr != "" {
		limitNum, err := strconv.Atoi(limitStr)
		if err != nil || limitNum < 1 || limitNum > 100 {
			validationErrs = append(validationErrs, errors.NewValidationError("limit", "Limit must be between 1 and 100", limitStr))
		} else {
			limit = limitNum
		}
	}

	var filters []string
	var args []interface{}

	if startDate != "" {
		startDate, err := utils.ParseDate(startDate)
		if err != nil {
			validationErrs = append(validationErrs, errors.NewValidationError("start_date", "Start date must be in YYYY-MM-DD format", startDate))
		}
		filters = append(filters, "r.date >= ?")
		args = append(args, startDate)
	}
	if endDate != "" {
		endDate, err := utils.ParseDate(endDate)
		if err != nil {
			validationErrs = append(validationErrs, errors.NewValidationError("end_date", "End date must be in YYYY-MM-DD format", endDate))
			errors.HandleValidationErrors(c, validationErrs)
			return
		}
		filters = append(filters, "r.date <= ?")
		args = append(args, endDate)
	}
	if category != "" {
		filters = append(filters, "c.name = ?")
		args = append(args, category)
	}
	if recordType != "" {
		filters = append(filters, "r.type = ?")
		args = append(args, recordType)
	}
	// Note: description filtering is done in Go after decryption
	if minAmountStr != "" {
		minAmount, err := strconv.ParseFloat(minAmountStr, 64)
		if err != nil {
			validationErrs = append(validationErrs, errors.NewValidationError("min_amount", "min_amount must be a number", minAmountStr))
		} else {
			filters = append(filters, "r.amount >= ?")
			args = append(args, minAmount)
		}
	}
	if maxAmountStr != "" {
		maxAmount, err := strconv.ParseFloat(maxAmountStr, 64)
		if err != nil {
			validationErrs = append(validationErrs, errors.NewValidationError("max_amount", "max_amount must be a number", maxAmountStr))
		} else {
			filters = append(filters, "r.amount <= ?")
			args = append(args, maxAmount)
		}
	}

	validationErrs = append(validationErrs, validator.GetErrors()...)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Calculate offset for pagination
	offset := (page - 1) * limit

	query := `
		SELECT r.id, r.date, r.description, c.name as category, r.amount, r.type, r.note, r.balance
		FROM records r
		JOIN categories c ON r.category_id = c.id
	`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY r.date DESC LIMIT ? OFFSET ?"

	// Add pagination parameters to args
	args = append(args, limit, offset)

	slog.Debug("Executing query", "query", query, "args", args)

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		appErr := errors.NewDatabase("Failed to retrieve records", err)
		errors.HandleError(c, appErr)
		return
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var rec models.Record
		err := rows.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Note, &rec.Balance)
		if err != nil {
			appErr := errors.NewDatabase("Failed to read record data", err)
			errors.HandleError(c, appErr)
			return
		}

		// Decrypt sensitive fields with proper error handling
		if rec.Description != "" {
			decrypted, err := secure.Decrypt(rec.Description)
			if err != nil {
				slog.Warn("Failed to decrypt description", "record_id", rec.ID, "error", err)
				rec.Description = "[Encryption Error]"
			} else {
				rec.Description = decrypted
			}
		}

		if rec.Note != "" {
			decrypted, err := secure.Decrypt(rec.Note)
			if err != nil {
				slog.Warn("Failed to decrypt Note", "record_id", rec.ID, "error", err)
				rec.Note = "[Encryption Error]"
			} else {
				rec.Note = decrypted
			}
		}

		// Filter by description substring after decryption
		if desc != "" {
			if !strings.Contains(strings.ToLower(rec.Description), strings.ToLower(desc)) {
				continue // skip this record if description does not match
			}
		}

		records = append(records, rec)
	}

	if err = rows.Err(); err != nil {
		appErr := errors.NewDatabase("Error iterating through records", err)
		errors.HandleError(c, appErr)
		return
	}

	// Get total count for pagination metadata
	var totalCount int
	countQuery := `
		SELECT COUNT(*)
		FROM records r
		JOIN categories c ON r.category_id = c.id
	`
	if len(filters) > 0 {
		countQuery += " WHERE " + strings.Join(filters, " AND ")
	}

	// Remove pagination args for count query
	countArgs := args[:len(args)-2]

	err = h.DB.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		slog.Warn("Failed to get total count", "error", err)
		totalCount = len(records) // fallback to current page count
	}

	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	slog.Info("Records retrieved successfully",
		"count", len(records),
		"filters_applied", len(filters) > 0 || desc != "",
		"page", page,
		"limit", limit,
		"total_count", totalCount,
		"total_pages", totalPages,
	)

	// Return response with pagination metadata
	response := gin.H{
		"records": records,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total_count": totalCount,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	}

	c.JSON(http.StatusOK, response)
}
