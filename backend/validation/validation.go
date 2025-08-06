package validation

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
)

// Validator provides validation methods
type Validator struct {
	errors errors.ValidationErrors
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		errors: make(errors.ValidationErrors, 0),
	}
}

// ValidateRecord validates a record model
func (v *Validator) ValidateRecord(record *models.Record) errors.ValidationErrors {
	v.errors = make(errors.ValidationErrors, 0)
	
	// Validate required fields
	v.required("date", record.Date, "Date is required")
	v.required("category", record.Category, "Category is required")
	v.required("type", record.Type, "Type is required")
	
	// Validate date format
	if record.Date != "" {
		v.dateFormat("date", record.Date, "Date must be in YYYY-MM-DD format")
	}
	
	// Validate amount
	v.positiveNumber("amount", record.Amount, "Amount must be a positive number")
	
	// Validate type
	if record.Type != "" {
		v.enum("type", record.Type, []string{"income", "expense", "transfer"}, "Type must be income, expense, or transfer")
	}
	
	// Validate description length
	if record.Description != "" {
		v.maxLength("description", record.Description, 255, "Description must be 255 characters or less")
	}
	
	// Validate notes length
	if record.Notes != "" {
		v.maxLength("notes", record.Notes, 1000, "Notes must be 1000 characters or less")
	}
	
	return v.errors
}

// ValidateCategory validates a category model
func (v *Validator) ValidateCategory(category *models.Category) errors.ValidationErrors {
	v.errors = make(errors.ValidationErrors, 0)
	
	// Validate required fields
	v.required("name", category.Name, "Category name is required")
	
	// Validate name length
	if category.Name != "" {
		v.minLength("name", category.Name, 1, "Category name must be at least 1 character")
		v.maxLength("name", category.Name, 50, "Category name must be 50 characters or less")
		v.pattern("name", category.Name, `^[a-zA-Z0-9\s\-_]+$`, "Category name contains invalid characters")
	}
	
	// Validate icon length
	if category.Icon != "" {
		v.maxLength("icon", category.Icon, 50, "Icon must be 50 characters or less")
	}
	
	// Validate color format (hex color)
	if category.Color != "" {
		v.pattern("color", category.Color, `^#[0-9A-Fa-f]{6}$`, "Color must be a valid hex color (e.g., #FF0000)")
	}
	
	return v.errors
}

// ValidateID validates an ID parameter
func (v *Validator) ValidateID(idStr string) (int, errors.ValidationErrors) {
	v.errors = make(errors.ValidationErrors, 0)
	
	if idStr == "" {
		v.errors = append(v.errors, errors.NewValidationError("id", "ID is required", idStr))
		return 0, v.errors
	}
	
	id, err := strconv.Atoi(idStr)
	if err != nil {
		v.errors = append(v.errors, errors.NewValidationError("id", "ID must be a valid integer", idStr))
		return 0, v.errors
	}
	
	if id <= 0 {
		v.errors = append(v.errors, errors.NewValidationError("id", "ID must be a positive integer", id))
		return 0, v.errors
	}
	
	return id, v.errors
}

// Validation helper methods
func (v *Validator) required(field, value, message string) {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, errors.NewValidationError(field, message, value))
	}
}

func (v *Validator) minLength(field, value string, min int, message string) {
	if len(value) < min {
		v.errors = append(v.errors, errors.NewValidationError(field, message, value))
	}
}

func (v *Validator) maxLength(field, value string, max int, message string) {
	if len(value) > max {
		v.errors = append(v.errors, errors.NewValidationError(field, message, value))
	}
}

func (v *Validator) positiveNumber(field string, value float64, message string) {
	if value <= 0 {
		v.errors = append(v.errors, errors.NewValidationError(field, message, value))
	}
}

func (v *Validator) enum(field, value string, allowed []string, message string) {
	for _, allowedValue := range allowed {
		if value == allowedValue {
			return
		}
	}
	v.errors = append(v.errors, errors.NewValidationError(field, message, value))
}

func (v *Validator) pattern(field, value, pattern, message string) {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.errors = append(v.errors, errors.NewValidationError(field, message, value))
	}
}

func (v *Validator) dateFormat(field, value, message string) {
	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		v.errors = append(v.errors, errors.NewValidationError(field, message, value))
	}
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors returns the validation errors
func (v *Validator) GetErrors() errors.ValidationErrors {
	return v.errors
} 