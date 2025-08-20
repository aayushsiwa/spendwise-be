package errors

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

// Error types for different scenarios
var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("not found")
	ErrDatabase     = errors.New("database error")
	ErrEncryption   = errors.New("encryption error")
	ErrValidation   = errors.New("validation error")
	ErrInternal     = errors.New("internal server error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrBadRequest   = errors.New("bad request")
)

// AppError represents a structured application error
type AppError struct {
	Type       string                 `json:"type"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Err        error                  `json:"-"`
	Context    map[string]interface{} `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(errType string, message string, statusCode int, err error) *AppError {
	return &AppError{
		Type:       errType,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Context:    make(map[string]interface{}),
	}
}

// WithDetails adds additional details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithContext adds context information for logging
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// Log logs the error with structured logging
func (e *AppError) Log() {
	attrs := []any{
		"type", e.Type,
		"message", e.Message,
		"status_code", e.StatusCode,
	}

	if e.Err != nil {
		attrs = append(attrs, "underlying_error", e.Err.Error())
	}

	for k, v := range e.Context {
		attrs = append(attrs, k, v)
	}

	slog.Error("application error", attrs...)
}

// Helper functions for common error types
func NewInvalidInput(message string, err error) *AppError {
	return New("invalid_input", message, http.StatusBadRequest, err)
}

func NewNotFound(message string, err error) *AppError {
	return New("not_found", message, http.StatusNotFound, err)
}

func NewDatabase(message string, err error) *AppError {
	return New("database_error", message, http.StatusInternalServerError, err)
}

func NewEncryption(message string, err error) *AppError {
	return New("encryption_error", message, http.StatusInternalServerError, err)
}

func NewValidation(message string, details map[string]interface{}) *AppError {
	return New("validation_error", message, http.StatusBadRequest, nil).WithDetails(details)
}

func NewInternal(message string, err error) *AppError {
	return New("internal_error", message, http.StatusInternalServerError, err)
}

func NewUnauthorized(message string, err error) *AppError {
	return New("unauthorized", message, http.StatusUnauthorized, err)
}

func NewForbidden(message string, err error) *AppError {
	return New("forbidden", message, http.StatusForbidden, err)
}

func NewConflict(message string, err error) *AppError {
	return New("conflict", message, http.StatusConflict, err)
}

// HandleError handles errors and sends appropriate HTTP responses
func HandleError(c *gin.Context, err error) {
	var appErr *AppError

	// Check if it's already an AppError
	if errors.As(err, &appErr) {
		appErr.Log()
		c.JSON(appErr.StatusCode, gin.H{
			"error": gin.H{
				"type":    appErr.Type,
				"message": appErr.Message,
				"details": appErr.Details,
			},
		})
		return
	}

	// Handle database errors
	if errors.Is(err, sql.ErrNoRows) {
		appErr = NewNotFound("Resource not found", err)
	} else if strings.Contains(err.Error(), "database") || strings.Contains(err.Error(), "sql") {
		appErr = NewDatabase("Database operation failed", err)
	} else {
		// Default to internal server error
		appErr = NewInternal("An unexpected error occurred", err)
	}

	// Add context for debugging
	appErr.WithContext("handler", getHandlerName())
	appErr.WithContext("method", c.Request.Method)
	appErr.WithContext("path", c.Request.URL.Path)

	appErr.Log()
	c.JSON(appErr.StatusCode, gin.H{
		"error": gin.H{
			"type":    appErr.Type,
			"message": appErr.Message,
		},
	})
}

// getHandlerName returns the name of the calling function for context
func getHandlerName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}

// ValidationError represents a validation error with field-specific details
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "no validation errors"
	}

	messages := make([]string, len(v))
	for i, err := range v {
		messages[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
	}
	return strings.Join(messages, "; ")
}

// ToMap converts validation errors to a map for JSON response
func (v ValidationErrors) ToMap() map[string]interface{} {
	details := make(map[string]interface{})
	for _, err := range v {
		details[err.Field] = gin.H{
			"message": err.Message,
			"value":   err.Value,
		}
	}
	return details
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string, value interface{}) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// HandleValidationErrors handles validation errors specifically
func HandleValidationErrors(c *gin.Context, validationErrs ValidationErrors) {
	appErr := NewValidation("Validation failed", validationErrs.ToMap())
	appErr.Log()
	c.JSON(appErr.StatusCode, gin.H{
		"error": gin.H{
			"type":    appErr.Type,
			"message": appErr.Message,
			"details": appErr.Details,
		},
	})
}
