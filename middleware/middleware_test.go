package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"aayushsiwa/expense-tracker/errors"

	"github.com/gin-gonic/gin"
)

func suppressLogs() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestSecurityHeaders(t *testing.T) {
	suppressLogs()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %s", w.Header().Get("X-Content-Type-Options"))
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected X-Frame-Options: DENY, got %s", w.Header().Get("X-Frame-Options"))
	}
	if w.Header().Get("X-XSS-Protection") != "1; mode=block" {
		t.Errorf("expected X-XSS-Protection: 1; mode=block, got %s", w.Header().Get("X-XSS-Protection"))
	}
}

func TestRequestLogger(t *testing.T) {
	suppressLogs()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		status int
	}{
		{name: "status 200", status: http.StatusOK},
		{name: "status 400", status: http.StatusBadRequest},
		{name: "status 500", status: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(RequestLogger())
			r.GET("/test", func(c *gin.Context) {
				c.Status(tt.status)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestErrorHandler(t *testing.T) {
	suppressLogs()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ErrorHandler())
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	var body map[string]map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	errObj, ok := body["error"]
	if !ok {
		t.Fatal("expected error key in response body")
	}
	if errObj["type"] != "panic" {
		t.Errorf("expected error type panic, got %s", errObj["type"])
	}
	if errObj["message"] != "Internal server error" {
		t.Errorf("expected message 'Internal server error', got %s", errObj["message"])
	}
}

func TestErrorHandler_UnknownError(t *testing.T) {
	suppressLogs()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ErrorHandler())
	r.GET("/panic-unknown", func(c *gin.Context) {
		panic(struct{ msg string }{"test struct panic"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/panic-unknown", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	var body map[string]map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	errObj, ok := body["error"]
	if !ok {
		t.Fatal("expected error key in response body")
	}
	if errObj["type"] != "panic" {
		t.Errorf("expected error type panic, got %s", errObj["type"])
	}
}

func TestValidationErrorHandler(t *testing.T) {
	suppressLogs()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ValidationErrorHandler())
	r.GET("/validate", func(c *gin.Context) {
		validationErrs := errors.ValidationErrors{
			errors.NewValidationError("amount", "is required", nil),
		}
		_ = c.Error(validationErrs)
		c.Status(http.StatusBadRequest)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/validate", nil)
	r.ServeHTTP(w, req)

	var body map[string]map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	errObj, ok := body["error"]
	if !ok {
		t.Fatal("expected error key in response body")
	}
	if errObj["type"] != "validation_error" {
		t.Errorf("expected error type validation_error, got %v", errObj["type"])
	}
}
