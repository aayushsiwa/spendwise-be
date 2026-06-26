package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/mocks"

	"github.com/gin-gonic/gin"
)

func TestCreateBudget(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "InvalidJSON",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ZeroAmount",
			body:       `{"amount":0,"month":6,"year":2026,"categoryID":"cat1"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Amount must be greater than 0",
		},
		{
			name:       "NegativeAmount",
			body:       `{"amount":-50,"month":6,"year":2026,"categoryID":"cat1"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Amount must be greater than 0",
		},
		{
			name: "ServiceError",
			body: `{"amount":100,"month":6,"year":2026,"categoryID":"cat1"}`,
			mock: &mocks.MockService{
				CreateBudgetErr: apperrors.NewDatabase("db error", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			body:       `{"amount":500,"month":6,"year":2026,"categoryID":"cat1"}`,
			wantStatus: http.StatusCreated,
			wantBody:   `"ID"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/budgets", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.CreateBudget(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
