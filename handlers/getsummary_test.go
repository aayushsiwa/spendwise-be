package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/mocks"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

func TestGetSummary(t *testing.T) {
	t.Parallel()

	summary := &models.Summary{
		TotalIncome:  1000.0,
		TotalExpense: 500.0,
		Net:          500.0,
		Opening:      0.0,
		Closing:      500.0,
		Incomes:      []models.CategoryDetail{},
		Expenses:     []models.CategoryDetail{},
	}

	summaryExplicit := &models.Summary{
		TotalIncome:  2000.0,
		TotalExpense: 1500.0,
		Net:          500.0,
	}

	tests := []struct {
		name       string
		query      string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name: "service error returns 500",
			mock: &mocks.MockService{
				GetSummaryErr: apperrors.NewDatabase("failed to get summary", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success with defaults",
			mock: &mocks.MockService{
				GetSummaryFn: func(_ context.Context, from, to, _, _ string) (*models.Summary, error) {
					if from == "" || to == "" {
						return nil, fmt.Errorf("expected non-empty from/to defaults")
					}
					return summary, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "summary",
		},
		{
			name:  "with explicit date range",
			query: "from=2024-01-01&to=2024-01-31",
			mock: &mocks.MockService{
				GetSummaryFn: func(_ context.Context, from, to, _, _ string) (*models.Summary, error) {
					if from != "2024-01-01" {
						t.Errorf("expected from=2024-01-01, got %q", from)
					}
					if to != "2024-01-31" {
						t.Errorf("expected to=2024-01-31, got %q", to)
					}
					return summaryExplicit, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "with category and type filters",
			query: "category=food&type=expense",
			mock: &mocks.MockService{
				GetSummaryFn: func(_ context.Context, _, _, categoryFilter, typeFilter string) (*models.Summary, error) {
					if categoryFilter != "food" {
						t.Errorf("expected category=food, got %q", categoryFilter)
					}
					if typeFilter != "expense" {
						t.Errorf("expected type=expense, got %q", typeFilter)
					}
					return &models.Summary{}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "response contains summary key",
			wantStatus: http.StatusOK,
			wantBody:   "summary",
		},
		{
			name: "default from is current month first day",
			mock: &mocks.MockService{
				GetSummaryFn: func(_ context.Context, from, _, _, _ string) (*models.Summary, error) {
					if !strings.HasSuffix(from, "-01") {
						t.Errorf("expected from to end with '-01' (first day of month), got %q", from)
					}
					return &models.Summary{}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			path := "/summary"
			if tt.query != "" {
				path += "?" + tt.query
			}
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.GetSummary(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
