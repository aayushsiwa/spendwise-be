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
	"gorm.io/gorm"
)

func TestUpdateBudget(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "MissingID",
			id:         "",
			body:       `{"amount":100}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "InvalidJSON",
			id:         "budget-123",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "MissingAmount",
			id:         "budget-123",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Amount is required",
		},
		{
			name:       "ZeroAmount",
			id:         "budget-123",
			body:       `{"amount":0}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Amount must be greater than 0",
		},
		{
			name:       "NegativeAmount",
			id:         "budget-123",
			body:       `{"amount":-10}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "Amount must be greater than 0",
		},
		{
			name: "NotFound",
			id:   "nonexistent",
			body: `{"amount":200}`,
			mock: &mocks.MockService{
				UpdateBudgetErr: gorm.ErrRecordNotFound,
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "Budget not found",
		},
		{
			name: "ServiceError",
			id:   "budget-123",
			body: `{"amount":200}`,
			mock: &mocks.MockService{
				UpdateBudgetErr: apperrors.NewDatabase("db error", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			id:         "budget-123",
			body:       `{"amount":200}`,
			wantStatus: http.StatusOK,
			wantBody:   "Budget updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPatch, "/budgets/"+tt.id, strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = []gin.Param{{Key: "id", Value: tt.id}}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.UpdateBudget(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
