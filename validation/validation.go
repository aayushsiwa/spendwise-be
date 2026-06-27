package validation

import (
	"regexp"
	"slices"
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

// ValidatePatchRecord validates only the fields present in a partial update request
func (v *Validator) ValidatePatchRecord(req *models.UpdateRecordRequest) errors.ValidationErrors {
	v.errors = make(errors.ValidationErrors, 0)

	if req.Date != nil {
		v.dateFormat("date", *req.Date, "Date must be in YYYY-MM-DD format")
	}
	if req.Amount != nil {
		v.positiveNumber("amount", *req.Amount, "Amount must be a positive number")
	}
	if req.Description != nil {
		v.maxLength("description", *req.Description, 255, "Description must be 255 characters or less")
	}
	if req.Note != nil {
		v.maxLength("note", *req.Note, 1000, "Note must be 1000 characters or less")
	}

	return v.errors
}

// ValidateRecord validates a record model
func (v *Validator) ValidateRecord(record *models.Record) errors.ValidationErrors {
	v.errors = make(errors.ValidationErrors, 0)

	// Validate required fields
	v.required("date", record.Date, "Date is required")
	v.required("category", record.Category, "Category is required")
	// v.required("type", record.Type, "Type is required")

	// Validate date format
	if record.Date != "" {
		v.dateFormat("date", record.Date, "Date must be in YYYY-MM-DD format")
	}

	// Validate amount
	v.positiveNumber("amount", record.Amount, "Amount must be a positive number")

	// Validate type
	// if record.Type != "" {
	// 	v.enum("type", record.Type, []string{"income", "expense", "transfer"}, "Type must be income, expense, or transfer")
	// }

	// Validate description length
	if record.Description != "" {
		v.maxLength("description", record.Description, 255, "Description must be 255 characters or less")
	}

	// Validate note length
	if record.Note != "" {
		v.maxLength("note", record.Note, 1000, "Note must be 1000 characters or less")
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

// ValidateBudget validates a budget model
func (v *Validator) ValidateBudget(budget *models.Budget) errors.ValidationErrors {
	v.errors = make(errors.ValidationErrors, 0)

	v.required("categoryID", budget.CategoryID, "Category ID is required")
	v.positiveNumber("amount", budget.Amount, "Amount must be greater than 0")

	return v.errors
}

// ValidateGoal validates a goal model
func (v *Validator) ValidateGoal(goal *models.Goal) errors.ValidationErrors {
	v.errors = make(errors.ValidationErrors, 0)

	v.required("name", goal.Name, "Goal name is required")
	if goal.Name != "" {
		v.maxLength("name", goal.Name, 100, "Goal name must be 100 characters or less")
	}
	v.positiveNumber("targetAmount", goal.TargetAmount, "Target amount must be greater than 0")

	if goal.TargetDate != "" {
		v.dateFormat("targetDate", goal.TargetDate, "Target date must be in YYYY-MM-DD format")
	}

	if goal.Status != "" {
		v.enum("status", string(goal.Status), []string{"active", "achieved", "abandoned"}, "Status must be active, achieved, or abandoned")
	}

	if goal.Description != "" {
		v.maxLength("description", goal.Description, 500, "Description must be 500 characters or less")
	}

	return v.errors
}

// ValidateUpdateGoal validates only the fields present in a goal update request
func (v *Validator) ValidateUpdateGoal(req *models.UpdateGoalRequest) errors.ValidationErrors {
	v.errors = make(errors.ValidationErrors, 0)

	if req.Name != nil {
		v.required("name", *req.Name, "Goal name cannot be empty")
		v.maxLength("name", *req.Name, 100, "Goal name must be 100 characters or less")
	}
	if req.TargetAmount != nil {
		v.positiveNumber("targetAmount", *req.TargetAmount, "Target amount must be greater than 0")
	}
	if req.CurrentAmount != nil {
		if *req.CurrentAmount < 0 {
			v.errors = append(v.errors, errors.NewValidationError("currentAmount", "Current amount must be non-negative", *req.CurrentAmount))
		}
	}
	if req.TargetDate != nil && *req.TargetDate != "" {
		v.dateFormat("targetDate", *req.TargetDate, "Target date must be in YYYY-MM-DD format")
	}
	if req.Status != nil {
		v.enum("status", *req.Status, []string{"active", "achieved", "abandoned"}, "Status must be active, achieved, or abandoned")
	}
	if req.Description != nil {
		v.maxLength("description", *req.Description, 500, "Description must be 500 characters or less")
	}

	return v.errors
}

// ValidateAddProgress validates a progress addition request
func (v *Validator) ValidateAddProgress(req *models.AddProgressRequest) errors.ValidationErrors {
	v.errors = make(errors.ValidationErrors, 0)

	if req == nil {
		v.errors = append(v.errors, errors.NewValidationError("amount", "Request body is required", nil))
		return v.errors
	}
	v.positiveNumber("amount", req.Amount, "Progress amount must be greater than 0")

	return v.errors
}

// ValidateUpdateBudgetAmount validates the amount field in a budget update request
func (v *Validator) ValidateUpdateBudgetAmount(amount *float64) errors.ValidationErrors {
	v.errors = make(errors.ValidationErrors, 0)

	if amount == nil {
		v.errors = append(v.errors, errors.NewValidationError("amount", "Amount is required", nil))
		return v.errors
	}
	v.positiveNumber("amount", *amount, "Amount must be greater than 0")

	return v.errors
}

// ValidateID validates a string ID parameter (records, categories, budgets)
func (v *Validator) ValidateID(idStr string) (string, errors.ValidationErrors) {
	v.errors = make(errors.ValidationErrors, 0)

	if idStr == "" || strings.TrimSpace(idStr) == "" {
		v.errors = append(v.errors, errors.NewValidationError("id", "ID is required", idStr))
		return "", v.errors
	}

	return idStr, v.errors
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
	if slices.Contains(allowed, value) {
		return
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
