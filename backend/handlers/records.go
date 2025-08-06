package handlers

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/secure"
	"aayushsiwa/expense-tracker/utils"
	"aayushsiwa/expense-tracker/validation"

	"github.com/gin-gonic/gin"
)

func GetRecords(c *gin.Context) {
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
		SELECT r.id, r.date, r.description, c.name as category, r.amount, r.type, r.note
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

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		appErr := errors.NewDatabase("Failed to retrieve records", err)
		errors.HandleError(c, appErr)
		return
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var rec models.Record
		err := rows.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Note)
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

	err = db.DB.QueryRow(countQuery, countArgs...).Scan(&totalCount)
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
		"data": records,
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

func GetRecord(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	row := db.DB.QueryRow(`
		SELECT r.id, r.date, r.description, c.name as category, r.amount, r.type, r.note
		FROM records r
		JOIN categories c ON r.category_id = c.id
		WHERE r.id = ?
	`, id)

	var rec models.Record
	if err := row.Scan(&rec.ID, &rec.Date, &rec.Description, &rec.Category, &rec.Amount, &rec.Type, &rec.Note); err != nil {
		if err == sql.ErrNoRows {
			appErr := errors.NewNotFound(fmt.Sprintf("Record with ID %d not found", id), err)
			errors.HandleError(c, appErr)
		} else {
			appErr := errors.NewDatabase("Failed to read record", err)
			errors.HandleError(c, appErr)
		}
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

	slog.Info("Record retrieved successfully", "record_id", rec.ID)
	c.JSON(http.StatusOK, rec)
}

func CreateRecord(c *gin.Context) {
	var rec models.Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		appErr := errors.NewInvalidInput("Invalid JSON body", err)
		errors.HandleError(c, appErr)
		return
	}

	// Validate record data
	validator := validation.NewValidator()
	validationErrs := validator.ValidateRecord(&rec)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Encrypt sensitive fields with proper error handling
	if rec.Description != "" {
		encrypted, err := secure.Encrypt(rec.Description)
		if err != nil {
			appErr := errors.NewEncryption("Failed to encrypt description", err)
			errors.HandleError(c, appErr)
			return
		}
		rec.Description = encrypted
	}

	if rec.Note != "" {
		encrypted, err := secure.Encrypt(rec.Note)
		if err != nil {
			appErr := errors.NewEncryption("Failed to encrypt note", err)
			errors.HandleError(c, appErr)
			return
		}
		rec.Note = encrypted
	}

	// Generate custom ID
	customId, err := utils.GenerateCustomID(rec.Date)
	if err != nil {
		appErr := errors.NewInternal("Failed to generate record ID", err)
		errors.HandleError(c, appErr)
		return
	}

	rec.ID, err = strconv.Atoi(customId)
	if err != nil {
		appErr := errors.NewInternal("Failed to parse generated ID", err)
		errors.HandleError(c, appErr)
		return
	}

	// Get category ID
	var categoryId int
	err = db.DB.QueryRow("SELECT id FROM categories WHERE name = ?", rec.Category).Scan(&categoryId)
	if err != nil {
		if err == sql.ErrNoRows {
			appErr := errors.NewInvalidInput("Category not found", err).WithDetails(map[string]interface{}{
				"category": rec.Category,
			})
			errors.HandleError(c, appErr)
		} else {
			appErr := errors.NewDatabase("Failed to find category", err)
			errors.HandleError(c, appErr)
		}
		return
	}

	// Insert record
	_, err = db.DB.Exec(`
		INSERT INTO records (id, date, description, category_id, amount, type, note)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.Date, rec.Description, categoryId, rec.Amount, rec.Type, rec.Note)
	if err != nil {
		appErr := errors.NewDatabase("Failed to insert record", err)
		errors.HandleError(c, appErr)
		return
	}

	// Update summary
	if err := UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record creation", "record_id", rec.ID, "error", err)
	}

	slog.Info("Record created successfully", "record_id", rec.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Record with id %d created successfully", rec.ID),
		"id":      rec.ID,
	})
}

func PatchRecord(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	var rec models.Record
	if err := c.ShouldBindJSON(&rec); err != nil {
		appErr := errors.NewInvalidInput("Invalid JSON body", err)
		errors.HandleError(c, appErr)
		return
	}

	// Validate record data
	validationErrs = validator.ValidateRecord(&rec)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Encrypt sensitive fields with proper error handling
	if rec.Description != "" {
		encrypted, err := secure.Encrypt(rec.Description)
		if err != nil {
			appErr := errors.NewEncryption("Failed to encrypt description", err)
			errors.HandleError(c, appErr)
			return
		}
		rec.Description = encrypted
	}

	if rec.Note != "" {
		encrypted, err := secure.Encrypt(rec.Note)
		if err != nil {
			appErr := errors.NewEncryption("Failed to encrypt note", err)
			errors.HandleError(c, appErr)
			return
		}
		rec.Note = encrypted
	}

	// Get category ID
	var categoryId int
	err := db.DB.QueryRow("SELECT id FROM categories WHERE name = ?", rec.Category).Scan(&categoryId)
	if err != nil {
		if err == sql.ErrNoRows {
			appErr := errors.NewInvalidInput("Category not found", err).WithDetails(map[string]interface{}{
				"category": rec.Category,
			})
			errors.HandleError(c, appErr)
		} else {
			appErr := errors.NewDatabase("Failed to find category", err)
			errors.HandleError(c, appErr)
		}
		return
	}

	// Check if record exists
	var exists int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM records WHERE id = ?", id).Scan(&exists)
	if err != nil {
		appErr := errors.NewDatabase("Failed to check record existence", err)
		errors.HandleError(c, appErr)
		return
	}

	if exists == 0 {
		appErr := errors.NewNotFound(fmt.Sprintf("Record with ID %d not found", id), nil)
		errors.HandleError(c, appErr)
		return
	}

	// Update record
	_, err = db.DB.Exec(`
		UPDATE records 
		SET date = ?, description = ?, category_id = ?, amount = ?, type = ?, note = ?
		WHERE id = ?`,
		rec.Date, rec.Description, categoryId, rec.Amount, rec.Type, rec.Note, id)
	if err != nil {
		appErr := errors.NewDatabase("Failed to update record", err)
		errors.HandleError(c, appErr)
		return
	}

	// Update summary
	if err := UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record update", "record_id", id, "error", err)
	}

	rec.ID = id
	slog.Info("Record updated successfully", "record_id", rec.ID)
	c.JSON(http.StatusOK, rec)
}

func DeleteRecord(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	res, err := db.DB.Exec(`DELETE FROM records WHERE id = ?`, id)
	if err != nil {
		appErr := errors.NewDatabase("Failed to delete record", err)
		errors.HandleError(c, appErr)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		appErr := errors.NewDatabase("Failed to get affected rows", err)
		errors.HandleError(c, appErr)
		return
	}

	if rowsAffected == 0 {
		appErr := errors.NewNotFound(fmt.Sprintf("Record with ID %d not found", id), nil)
		errors.HandleError(c, appErr)
		return
	}

	// Update summary
	if err := UpdateSummary(); err != nil {
		slog.Warn("Failed to update summary after record deletion", "record_id", id, "error", err)
	}

	slog.Info("Record deleted successfully", "record_id", id)
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Record with id %d deleted successfully", id),
	})
}
