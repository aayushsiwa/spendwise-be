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

func TestDeleteBudget(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "MissingID",
			id:         "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "NotFound",
			id:   "nonexistent",
			mock: &mocks.MockService{
				DeleteBudgetErr: gorm.ErrRecordNotFound,
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "Budget not found",
		},
		{
			name: "ServiceError",
			id:   "budget-123",
			mock: &mocks.MockService{
				DeleteBudgetErr: apperrors.NewDatabase("db error", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			id:         "budget-123",
			wantStatus: http.StatusOK,
			wantBody:   "Budget deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/budgets/"+tt.id, nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.id}}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.DeleteBudget(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
