# Error Handling System

This document describes the comprehensive error handling system implemented in the Expense Tracker backend.

## Overview

The error handling system provides:
- **Structured error responses** with consistent JSON format
- **Centralized error handling** with proper HTTP status codes
- **Detailed validation errors** with field-specific messages
- **Structured logging** for debugging and monitoring
- **Graceful error recovery** with panic handling
- **Security-focused error messages** that don't leak sensitive information

## Error Response Format

All error responses follow this consistent JSON structure:

```json
{
  "error": {
    "type": "error_type",
    "message": "Human-readable error message",
    "details": {
      "field_name": {
        "message": "Field-specific error message",
        "value": "Invalid value provided"
      }
    }
  }
}
```

## Error Types

### 1. Application Errors (`errors.AppError`)

Custom application errors with structured information:

```go
type AppError struct {
    Type       string                 // Error type identifier
    Message    string                 // Human-readable message
    Details    map[string]interface{} // Additional error details
    StatusCode int                    // HTTP status code
    Err        error                  // Underlying error
    Context    map[string]interface{} // Debug context
}
```

### 2. Validation Errors (`errors.ValidationErrors`)

Collection of field-specific validation errors:

```go
type ValidationError struct {
    Field   string      // Field name
    Message string      // Error message
    Value   interface{} // Invalid value
}
```

## Error Categories

### Database Errors
- **Type**: `database_error`
- **Status Code**: 500
- **Use Case**: SQL errors, connection issues, transaction failures

### Validation Errors
- **Type**: `validation_error`
- **Status Code**: 400
- **Use Case**: Invalid input data, missing required fields

### Not Found Errors
- **Type**: `not_found`
- **Status Code**: 404
- **Use Case**: Resource doesn't exist

### Invalid Input Errors
- **Type**: `invalid_input`
- **Status Code**: 400
- **Use Case**: Malformed JSON, invalid parameters

### Encryption Errors
- **Type**: `encryption_error`
- **Status Code**: 500
- **Use Case**: Encryption/decryption failures

### Conflict Errors
- **Type**: `conflict`
- **Status Code**: 409
- **Use Case**: Resource conflicts, business rule violations

## Usage Examples

### Creating Custom Errors

```go
// Simple error
appErr := errors.NewDatabase("Failed to connect to database", err)

// Error with details
appErr := errors.NewNotFound("User not found", err).WithDetails(map[string]interface{}{
    "user_id": userId,
})

// Error with context for debugging
appErr := errors.NewInternal("Processing failed", err).
    WithContext("user_id", userId).
    WithContext("operation", "data_processing")
```

### Handling Errors in Handlers

```go
func GetUser(c *gin.Context) {
    user, err := db.GetUser(id)
    if err != nil {
        errors.HandleError(c, err)
        return
    }
    c.JSON(http.StatusOK, user)
}
```

### Validation Error Handling

```go
func CreateUser(c *gin.Context) {
    var user models.User
    if err := c.ShouldBindJSON(&user); err != nil {
        appErr := errors.NewInvalidInput("Invalid JSON body", err)
        errors.HandleError(c, appErr)
        return
    }

    validator := validation.NewValidator()
    validationErrs := validator.ValidateUser(&user)
    if len(validationErrs) > 0 {
        errors.HandleValidationErrors(c, validationErrs)
        return
    }
    
    // Process valid user...
}
```

## Middleware

### Error Handler Middleware
- **Purpose**: Panic recovery and centralized error handling
- **Features**: Stack trace logging, structured error responses

### Request Logger Middleware
- **Purpose**: Structured request/response logging
- **Features**: Latency tracking, status code logging, error detection

### Rate Limiter Middleware
- **Purpose**: Prevent abuse and ensure fair usage
- **Features**: IP-based rate limiting (100 requests/minute)

### Security Headers Middleware
- **Purpose**: Add security headers to all responses
- **Features**: XSS protection, content type options, frame options

## Validation System

### Input Validation
The validation system provides comprehensive input validation:

```go
// Record validation
validationErrs := validator.ValidateRecord(&record)

// Category validation
validationErrs := validator.ValidateCategory(&category)

// ID validation
id, validationErrs := validator.ValidateID(idStr)
```

### Validation Rules
- **Required fields**: Non-empty strings, positive numbers
- **Format validation**: Date formats, email patterns, hex colors
- **Length limits**: String length constraints
- **Enum validation**: Allowed values for specific fields
- **Pattern matching**: Regex-based validation

## Logging

### Structured Logging
All errors are logged with structured information:

```json
{
  "level": "error",
  "msg": "application error",
  "type": "database_error",
  "message": "Failed to connect to database",
  "status_code": 500,
  "underlying_error": "connection refused",
  "handler": "handlers.GetUser",
  "method": "GET",
  "path": "/api/v1/users/123"
}
```

### Log Levels
- **ERROR**: Application errors, panics, critical failures
- **WARN**: Non-critical issues, rate limiting, validation warnings
- **INFO**: Successful operations, important events
- **DEBUG**: Detailed debugging information

## Best Practices

### 1. Always Handle Errors
```go
// Good
if err != nil {
    errors.HandleError(c, err)
    return
}

// Bad
if err != nil {
    log.Println(err) // Don't just log and continue
}
```

### 2. Use Appropriate Error Types
```go
// Good - specific error type
if err == sql.ErrNoRows {
    appErr := errors.NewNotFound("User not found", err)
}

// Bad - generic error
c.JSON(500, gin.H{"error": "Something went wrong"})
```

### 3. Provide Context for Debugging
```go
// Good - with context
appErr := errors.NewDatabase("Failed to update user", err).
    WithContext("user_id", userId).
    WithContext("operation", "update_profile")

// Bad - no context
appErr := errors.NewDatabase("Database error", err)
```

### 4. Validate Input Early
```go
// Good - validate before processing
validationErrs := validator.ValidateRecord(&record)
if len(validationErrs) > 0 {
    errors.HandleValidationErrors(c, validationErrs)
    return
}

// Bad - validate after processing
// Process record...
// Then check if it's valid
```

### 5. Use Safe Error Messages
```go
// Good - safe for production
appErr := errors.NewInternal("An unexpected error occurred", err)

// Bad - might leak sensitive info
appErr := errors.NewInternal("Database password is wrong", err)
```

## Security Considerations

### Error Information Disclosure
- Never expose internal error details in production
- Log sensitive information only at DEBUG level
- Use generic error messages for external users

### Input Validation
- Validate all user inputs
- Use parameterized queries to prevent SQL injection
- Sanitize error messages to prevent XSS

### Rate Limiting
- Implement rate limiting to prevent abuse
- Monitor for unusual request patterns
- Log rate limit violations for analysis

This error handling system provides a robust foundation for building reliable, maintainable, and secure APIs. 