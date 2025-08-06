package handlers

import (
	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/validation"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetCategories(c *gin.Context) {
	rows, err := db.DB.Query("SELECT id, name, icon, color FROM categories ORDER BY name ASC")
	if err != nil {
		appErr := errors.NewDatabase("Failed to fetch categories", err)
		errors.HandleError(c, appErr)
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Icon, &cat.Color); err != nil {
			slog.Warn("Failed to scan category row", "error", err)
			continue // Skip invalid rows but continue processing
		}
		categories = append(categories, cat)
	}

	if err = rows.Err(); err != nil {
		appErr := errors.NewDatabase("Error iterating through categories", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Categories retrieved successfully", "count", len(categories))
	c.JSON(http.StatusOK, categories)
}

func GetCategoryRecords(c *gin.Context) {
	pathParts := strings.Split(c.Request.URL.Path, "/")
	if len(pathParts) < 2 {
		appErr := errors.NewInvalidInput("Invalid URL path", nil)
		errors.HandleError(c, appErr)
		return
	}
	categoryId := pathParts[len(pathParts)-1]

	// Validate category ID
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(categoryId)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	var categoryName string
	err := db.DB.QueryRow("SELECT name FROM categories WHERE id = ?", id).Scan(&categoryName)
	if err != nil {
		appErr := errors.NewNotFound("Category not found", err).WithDetails(map[string]interface{}{
			"category_id": id,
		})
		errors.HandleError(c, appErr)
		return
	}

	rows, err := db.DB.Query(`
		SELECT r.id, r.date, r.description, r.category_id, c.name, r.amount, r.type, r.note
		FROM records r
		JOIN categories c ON r.category_id = c.id
		WHERE c.name = ?
		ORDER BY r.date DESC
	`, categoryName)
	if err != nil {
		appErr := errors.NewDatabase("Failed to fetch records", err)
		errors.HandleError(c, appErr)
		return
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var r models.Record
		err := rows.Scan(&r.ID, &r.Date, &r.Description, &categoryId, &r.Category, &r.Amount, &r.Type, &r.Note)
		if err != nil {
			slog.Warn("Failed to scan record row", "error", err)
			continue // Skip invalid rows but continue processing
		}
		records = append(records, r)
	}

	if err = rows.Err(); err != nil {
		appErr := errors.NewDatabase("Error iterating through records", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Category records retrieved successfully", "category", categoryName, "count", len(records))
	c.JSON(http.StatusOK, records)
}

func CreateCategories(c *gin.Context) {
	var categories []models.Category

	if err := c.BindJSON(&categories); err != nil {
		appErr := errors.NewInvalidInput("Invalid JSON body", err)
		errors.HandleError(c, appErr)
		return
	}

	if len(categories) == 0 {
		appErr := errors.NewInvalidInput("No categories provided", nil)
		errors.HandleError(c, appErr)
		return
	}

	// Validate all categories before processing
	validator := validation.NewValidator()
	var allValidationErrs errors.ValidationErrors

	for i, cat := range categories {
		validationErrs := validator.ValidateCategory(&cat)
		for _, err := range validationErrs {
			// Add index to field name for better error reporting
			err.Field = fmt.Sprintf("categories[%d].%s", i, err.Field)
			allValidationErrs = append(allValidationErrs, err)
		}
	}

	if len(allValidationErrs) > 0 {
		errors.HandleValidationErrors(c, allValidationErrs)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		appErr := errors.NewDatabase("Failed to begin transaction", err)
		errors.HandleError(c, appErr)
		return
	}

	stmt, err := tx.Prepare("INSERT INTO categories (name, icon, color) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		appErr := errors.NewDatabase("Failed to prepare statement", err)
		errors.HandleError(c, appErr)
		return
	}
	defer stmt.Close()

	var inserted []gin.H
	for _, cat := range categories {
		if cat.Name == "" {
			continue
		}
		lowerName := strings.ToLower(cat.Name)
		result, err := stmt.Exec(lowerName, cat.Icon, cat.Color)
		if err != nil {
			tx.Rollback()
			appErr := errors.NewDatabase("Failed to insert category", err).WithDetails(map[string]interface{}{
				"category_name": cat.Name,
			})
			errors.HandleError(c, appErr)
			return
		}
		id, _ := result.LastInsertId()
		inserted = append(inserted, gin.H{
			"id":    id,
			"name":  lowerName,
			"icon":  cat.Icon,
			"color": cat.Color,
		})
	}

	if err := tx.Commit(); err != nil {
		appErr := errors.NewDatabase("Failed to commit transaction", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Categories created successfully", "count", len(inserted))
	c.JSON(http.StatusCreated, inserted)
}

func UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	var cat models.Category
	if err := c.BindJSON(&cat); err != nil {
		appErr := errors.NewInvalidInput("Invalid JSON body", err)
		errors.HandleError(c, appErr)
		return
	}

	// Validate category data
	validationErrs = validator.ValidateCategory(&cat)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Check if category exists
	var exists int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE id = ?", id).Scan(&exists)
	if err != nil {
		appErr := errors.NewDatabase("Failed to check category existence", err)
		errors.HandleError(c, appErr)
		return
	}

	if exists == 0 {
		appErr := errors.NewNotFound("Category not found", nil).WithDetails(map[string]interface{}{
			"category_id": id,
		})
		errors.HandleError(c, appErr)
		return
	}

	_, err = db.DB.Exec("UPDATE categories SET name = ?, icon = ?, color = ? WHERE id = ?",
		cat.Name, cat.Icon, cat.Color, id)
	if err != nil {
		appErr := errors.NewDatabase("Failed to update category", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Category updated successfully", "category_id", id, "name", cat.Name)
	c.JSON(http.StatusOK, gin.H{"id": id, "name": cat.Name, "icon": cat.Icon, "color": cat.Color})
}

func DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")

	// Validate ID parameter
	validator := validation.NewValidator()
	id, validationErrs := validator.ValidateID(idStr)
	if len(validationErrs) > 0 {
		errors.HandleValidationErrors(c, validationErrs)
		return
	}

	// Check if category exists
	var exists int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM categories WHERE id = ?", id).Scan(&exists)
	if err != nil {
		appErr := errors.NewDatabase("Failed to check category existence", err)
		errors.HandleError(c, appErr)
		return
	}

	if exists == 0 {
		appErr := errors.NewNotFound("Category not found", nil).WithDetails(map[string]interface{}{
			"category_id": id,
		})
		errors.HandleError(c, appErr)
		return
	}

	// Check if category is being used by any records
	var recordCount int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM records WHERE category_id = ?", id).Scan(&recordCount)
	if err != nil {
		appErr := errors.NewDatabase("Failed to check category usage", err)
		errors.HandleError(c, appErr)
		return
	}

	if recordCount > 0 {
		appErr := errors.NewConflict("Cannot delete category that has associated records", nil).WithDetails(map[string]interface{}{
			"category_id":  id,
			"record_count": recordCount,
		})
		errors.HandleError(c, appErr)
		return
	}

	_, err = db.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		appErr := errors.NewDatabase("Failed to delete category", err)
		errors.HandleError(c, appErr)
		return
	}

	slog.Info("Category deleted successfully", "category_id", id)
	c.JSON(http.StatusNoContent, nil)
}
