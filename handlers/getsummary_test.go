package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

func TestGetSummary_ServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/summary", nil)

	svc := &mockService{
		getSummaryFn: func(_ context.Context, _, _, _, _ string) (*models.Summary, error) {
			return nil, apperrors.NewDatabase("failed to get summary", nil)
		},
	}
	h := &Handler{Service: svc}
	h.GetSummary(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetSummary_SuccessWithDefaults(t *testing.T) {
	summary := &models.Summary{
		TotalIncome:  1000.0,
		TotalExpense: 500.0,
		Net:          500.0,
		Opening:      0.0,
		Closing:      500.0,
		Incomes:      []models.CategoryDetail{},
		Expenses:     []models.CategoryDetail{},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/summary", nil)

	svc := &mockService{
		getSummaryFn: func(_ context.Context, from, to, categoryFilter, typeFilter string) (*models.Summary, error) {
			// from and to should be defaulted to current month
			if from == "" || to == "" {
				return nil, fmt.Errorf("expected non-empty from/to defaults")
			}
			return summary, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetSummary(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response: %v", err)
	}
	if _, ok := resp["summary"]; !ok {
		t.Error("expected 'summary' key in response")
	}
}

func TestGetSummary_WithExplicitDateRange(t *testing.T) {
	summary := &models.Summary{
		TotalIncome:  2000.0,
		TotalExpense: 1500.0,
		Net:          500.0,
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/summary?from=2024-01-01&to=2024-01-31", nil)

	var capturedFrom, capturedTo string
	svc := &mockService{
		getSummaryFn: func(_ context.Context, from, to, _, _ string) (*models.Summary, error) {
			capturedFrom = from
			capturedTo = to
			return summary, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetSummary(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedFrom != "2024-01-01" {
		t.Errorf("expected from=2024-01-01, got %q", capturedFrom)
	}
	if capturedTo != "2024-01-31" {
		t.Errorf("expected to=2024-01-31, got %q", capturedTo)
	}
}

func TestGetSummary_WithCategoryAndTypeFilters(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/summary?category=food&type=expense", nil)

	var capturedCat, capturedType string
	svc := &mockService{
		getSummaryFn: func(_ context.Context, _, _, categoryFilter, typeFilter string) (*models.Summary, error) {
			capturedCat = categoryFilter
			capturedType = typeFilter
			return &models.Summary{}, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetSummary(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if capturedCat != "food" {
		t.Errorf("expected category=food, got %q", capturedCat)
	}
	if capturedType != "expense" {
		t.Errorf("expected type=expense, got %q", capturedType)
	}
}

func TestGetSummary_ResponseContainsSummaryKey(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/summary", nil)

	h := &Handler{Service: &mockService{}}
	h.GetSummary(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"summary"`) {
		t.Errorf("expected 'summary' key in response body, got: %s", body)
	}
}

func TestGetSummary_DefaultFromIsCurrentMonthFirstDay(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/summary", nil)

	var capturedFrom string
	svc := &mockService{
		getSummaryFn: func(_ context.Context, from, _, _, _ string) (*models.Summary, error) {
			capturedFrom = from
			return &models.Summary{}, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetSummary(c)

	// Should end with "-01" (first day of month)
	if !strings.HasSuffix(capturedFrom, "-01") {
		t.Errorf("expected from to end with '-01' (first day of month), got %q", capturedFrom)
	}
}