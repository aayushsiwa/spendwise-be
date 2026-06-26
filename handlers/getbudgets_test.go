package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/mocks"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

func TestGetBudgets(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name: "ServiceError",
			mock: &mocks.MockService{
				GetBudgetsErr: apperrors.NewDatabase("db error", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "Success",
			mock: &mocks.MockService{
				GetBudgetsResult: []models.Budget{
					{ID: "b1", CategoryID: "cat1", Amount: 500, Month: 6, Year: 2026},
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"budgets"`,
		},
		{
			name:       "Empty",
			wantStatus: http.StatusOK,
			wantBody:   `"budgets"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/budgets", nil)

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.GetBudgets(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

// TestGetBudgetsPassesMonthYearParams verifies that query params month and year
// are parsed and forwarded to the service correctly.
func TestGetBudgetsPassesMonthYearParams(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantMonth  int
		wantYear   int
	}{
		{
			name:      "explicit month and year",
			query:     "/budgets?month=3&year=2025",
			wantMonth: 3,
			wantYear:  2025,
		},
		{
			name:      "only year provided uses given year",
			query:     "/budgets?year=2024",
			wantYear:  2024,
		},
		{
			name:      "invalid month falls back to zero (strconv.Atoi error ignored)",
			query:     "/budgets?month=notanumber&year=2025",
			wantMonth: 0,
			wantYear:  2025,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedMonth, capturedYear int
			svc := &mocks.MockService{
				GetBudgetsFn: func(_ context.Context, month, year int) ([]models.Budget, error) {
					capturedMonth = month
					capturedYear = year
					return nil, nil
				},
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.query, nil)

			h := &Handler{Service: svc}
			h.GetBudgets(c)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
			}
			if tt.wantMonth != 0 && capturedMonth != tt.wantMonth {
				t.Errorf("month = %d, want %d", capturedMonth, tt.wantMonth)
			}
			if tt.wantYear != 0 && capturedYear != tt.wantYear {
				t.Errorf("year = %d, want %d", capturedYear, tt.wantYear)
			}
		})
	}
}

func TestGetBudgetProgress(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name: "ServiceError",
			mock: &mocks.MockService{
				GetBudgetProgressErr: apperrors.NewDatabase("db error", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "Success",
			mock: &mocks.MockService{
				GetBudgetProgressResult: []models.BudgetProgress{
					{Budget: models.Budget{ID: "b1", Amount: 500}, Spent: 300, Percentage: 60},
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"progress"`,
		},
		{
			name:       "Empty",
			wantStatus: http.StatusOK,
			wantBody:   `"progress"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/budgets/progress", nil)

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.GetBudgetProgress(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

// TestGetBudgetProgressPassesMonthYearParams verifies that query params month and year
// are parsed and forwarded to the service correctly.
func TestGetBudgetProgressPassesMonthYearParams(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantMonth int
		wantYear  int
	}{
		{
			name:      "explicit month and year",
			query:     "/budgets/progress?month=11&year=2025",
			wantMonth: 11,
			wantYear:  2025,
		},
		{
			name:      "only month provided",
			query:     "/budgets/progress?month=1",
			wantMonth: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedMonth, capturedYear int
			svc := &mocks.MockService{
				GetBudgetProgressFn: func(_ context.Context, month, year int) ([]models.BudgetProgress, error) {
					capturedMonth = month
					capturedYear = year
					return nil, nil
				},
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.query, nil)

			h := &Handler{Service: svc}
			h.GetBudgetProgress(c)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
			}
			if tt.wantMonth != 0 && capturedMonth != tt.wantMonth {
				t.Errorf("month = %d, want %d", capturedMonth, tt.wantMonth)
			}
			if tt.wantYear != 0 && capturedYear != tt.wantYear {
				t.Errorf("year = %d, want %d", capturedYear, tt.wantYear)
			}
		})
	}
}
