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
