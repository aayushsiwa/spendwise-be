package errors

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestNewInvalidInput(t *testing.T) {
	err := NewInvalidInput("bad data", nil)
	if err.Type != "invalid_input" {
		t.Errorf("Type = %q, want %q", err.Type, "invalid_input")
	}
	if err.Message != "bad data" {
		t.Errorf("Message = %q, want %q", err.Message, "bad data")
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusBadRequest)
	}
}

func TestNewNotFound(t *testing.T) {
	err := NewNotFound("missing", sql.ErrNoRows)
	if err.Type != "not_found" {
		t.Errorf("Type = %q, want %q", err.Type, "not_found")
	}
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusNotFound)
	}
	if err.Err != sql.ErrNoRows {
		t.Error("Err should be sql.ErrNoRows")
	}
}

func TestNewDatabase(t *testing.T) {
	err := NewDatabase("db down", errors.New("connection refused"))
	if err.Type != "database_error" {
		t.Errorf("Type = %q, want %q", err.Type, "database_error")
	}
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusInternalServerError)
	}
}

func TestNewEncryption(t *testing.T) {
	err := NewEncryption("crypto fail", errors.New("bad key"))
	if err.Type != "encryption_error" {
		t.Errorf("Type = %q, want %q", err.Type, "encryption_error")
	}
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusInternalServerError)
	}
}

func TestNewValidation(t *testing.T) {
	details := map[string]any{"field": "age"}
	err := NewValidation("invalid", details)
	if err.Type != "validation_error" {
		t.Errorf("Type = %q, want %q", err.Type, "validation_error")
	}
	if err.Message != "invalid" {
		t.Errorf("Message = %q, want %q", err.Message, "invalid")
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusBadRequest)
	}
	if err.Err != nil {
		t.Error("Err should be nil for NewValidation")
	}
	if err.Details == nil || err.Details["field"] != "age" {
		t.Error("Details should contain the supplied map")
	}
}

func TestNewInternal(t *testing.T) {
	err := NewInternal("oops", errors.New("panic"))
	if err.Type != "internal_error" {
		t.Errorf("Type = %q, want %q", err.Type, "internal_error")
	}
	if err.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusInternalServerError)
	}
}

func TestNewUnauthorized(t *testing.T) {
	err := NewUnauthorized("login required", nil)
	if err.Type != "unauthorized" {
		t.Errorf("Type = %q, want %q", err.Type, "unauthorized")
	}
	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusUnauthorized)
	}
}

func TestNewForbidden(t *testing.T) {
	err := NewForbidden("no access", nil)
	if err.Type != "forbidden" {
		t.Errorf("Type = %q, want %q", err.Type, "forbidden")
	}
	if err.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusForbidden)
	}
}

func TestNewConflict(t *testing.T) {
	err := NewConflict("duplicate", nil)
	if err.Type != "conflict" {
		t.Errorf("Type = %q, want %q", err.Type, "conflict")
	}
	if err.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %d, want %d", err.StatusCode, http.StatusConflict)
	}
}

func TestAppError_Error(t *testing.T) {
	t.Run("with underlying error", func(t *testing.T) {
		e := New("test", "something failed", 500, errors.New("cause"))
		want := "something failed: cause"
		if got := e.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("without underlying error", func(t *testing.T) {
		e := New("test", "just a message", 400, nil)
		want := "just a message"
		if got := e.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestAppError_Unwrap(t *testing.T) {
	t.Run("returns nil when no underlying error", func(t *testing.T) {
		e := New("test", "msg", 500, nil)
		if got := e.Unwrap(); got != nil {
			t.Errorf("Unwrap() = %v, want nil", got)
		}
	})

	t.Run("returns underlying error", func(t *testing.T) {
		cause := errors.New("root cause")
		e := New("test", "msg", 500, cause)
		if got := e.Unwrap(); got != cause {
			t.Errorf("Unwrap() = %v, want %v", got, cause)
		}
	})
}

func TestAppError_WithDetails(t *testing.T) {
	e := New("test", "msg", 500, nil)
	details := map[string]any{"key": "val"}
	got := e.WithDetails(details)
	if got != e {
		t.Error("WithDetails should return the same pointer")
	}
	if e.Details["key"] != "val" {
		t.Errorf("Details = %v, want %v", e.Details, details)
	}
}

func TestAppError_WithContext(t *testing.T) {
	e := New("test", "msg", 500, nil)
	e.WithContext("req_id", "123")
	if e.Context["req_id"] != "123" {
		t.Errorf("Context[req_id] = %v, want %v", e.Context["req_id"], "123")
	}
}

func TestNewValidationError(t *testing.T) {
	ve := NewValidationError("email", "invalid format", "bad@")
	if ve.Field != "email" {
		t.Errorf("Field = %q, want %q", ve.Field, "email")
	}
	if ve.Message != "invalid format" {
		t.Errorf("Message = %q, want %q", ve.Message, "invalid format")
	}
	if ve.Value != "bad@" {
		t.Errorf("Value = %v, want %v", ve.Value, "bad@")
	}
}

func TestValidationErrors_Error(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		v := ValidationErrors{}
		if got := v.Error(); got != "no validation errors" {
			t.Errorf("Error() = %q, want %q", got, "no validation errors")
		}
	})

	t.Run("populated", func(t *testing.T) {
		v := ValidationErrors{
			{Field: "name", Message: "required"},
			{Field: "age", Message: "must be positive"},
		}
		want := "name: required; age: must be positive"
		if got := v.Error(); got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestValidationErrors_ToMap(t *testing.T) {
	v := ValidationErrors{
		{Field: "email", Message: "invalid", Value: "x"},
		{Field: "age", Message: "too low", Value: -1},
	}
	m := v.ToMap()
	if len(m) != 2 {
		t.Fatalf("ToMap() len = %d, want 2", len(m))
	}
	for _, field := range []string{"email", "age"} {
		entry, ok := m[field].(gin.H)
		if !ok {
			t.Errorf("ToMap()[%q] type = %T, want gin.H", field, m[field])
			continue
		}
		if entry["message"] == "" {
			t.Errorf("ToMap()[%q][message] is empty", field)
		}
	}
}

func TestHandleError_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	appErr := NewInvalidInput("bad request", errors.New("parse failed"))
	HandleError(c, appErr)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errorObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("response missing 'error' key")
	}
	if errorObj["type"] != "invalid_input" {
		t.Errorf("error.type = %v, want invalid_input", errorObj["type"])
	}
	if errorObj["message"] != "bad request" {
		t.Errorf("error.message = %v, want 'bad request'", errorObj["message"])
	}
}

func TestHandleError_SqlErrNoRows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/resource", nil)

	HandleError(c, sql.ErrNoRows)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errorObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("response missing 'error' key")
	}
	if errorObj["type"] != "not_found" {
		t.Errorf("error.type = %v, want not_found", errorObj["type"])
	}
	if errorObj["message"] != "Resource not found" {
		t.Errorf("error.message = %v, want 'Resource not found'", errorObj["message"])
	}
}

func TestHandleError_PlainError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	HandleError(c, errors.New("something went wrong"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errorObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("response missing 'error' key")
	}
	if errorObj["type"] != "internal_error" {
		t.Errorf("error.type = %v, want internal_error", errorObj["type"])
	}
	if errorObj["message"] != "An unexpected error occurred" {
		t.Errorf("error.message = %v, want 'An unexpected error occurred'", errorObj["message"])
	}
}

func TestHandleValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	ves := ValidationErrors{
		{Field: "name", Message: "required", Value: ""},
		{Field: "age", Message: "must be a number", Value: "abc"},
	}
	HandleValidationErrors(c, ves)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errorObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("response missing 'error' key")
	}
	if errorObj["type"] != "validation_error" {
		t.Errorf("error.type = %v, want validation_error", errorObj["type"])
	}
	if errorObj["message"] != "Validation failed" {
		t.Errorf("error.message = %v, want 'Validation failed'", errorObj["message"])
	}
	details, ok := errorObj["details"].(map[string]any)
	if !ok {
		t.Fatal("response missing error.details")
	}
	if _, exists := details["name"]; !exists {
		t.Error("details should contain 'name'")
	}
	if _, exists := details["age"]; !exists {
		t.Error("details should contain 'age'")
	}
}

func TestHandleError_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	HandleError(c, errors.New("sql: table not found"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	errorObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("response missing 'error' key")
	}
	if errorObj["type"] != "database_error" {
		t.Errorf("error.type = %v, want database_error", errorObj["type"])
	}
	if errorObj["message"] != "Database operation failed" {
		t.Errorf("error.message = %v, want 'Database operation failed'", errorObj["message"])
	}
}

func TestGetHandlerName_Mocked(t *testing.T) {
	// Save originals
	origCaller := runtimeCaller
	origFuncForPC := runtimeFuncForPC
	defer func() {
		runtimeCaller = origCaller
		runtimeFuncForPC = origFuncForPC
	}()

	t.Run("caller returns not ok", func(t *testing.T) {
		runtimeCaller = func(skip int) (pc uintptr, file string, line int, ok bool) {
			return 0, "", 0, false
		}
		if got := getHandlerName(); got != "unknown" {
			t.Errorf("getHandlerName() = %q, want %q", got, "unknown")
		}
	})

	t.Run("funcForPC returns nil", func(t *testing.T) {
		runtimeCaller = origCaller
		runtimeFuncForPC = func(pc uintptr) funcNameGetter {
			return nil
		}
		if got := getHandlerName(); got != "unknown" {
			t.Errorf("getHandlerName() = %q, want %q", got, "unknown")
		}
	})
}
