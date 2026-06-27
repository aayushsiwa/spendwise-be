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

func TestAddGoalProgress(t *testing.T) {
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
			id:         "goal-123",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_ZeroAmount",
			id:         "goal-123",
			body:       `{"amount":0}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_NegativeAmount",
			id:         "goal-123",
			body:       `{"amount":-50}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			id:         "nonexistent",
			body:       `{"amount":100}`,
			mock:       &mocks.MockService{AddGoalProgressErr: apperrors.NewNotFound("Goal not found", nil)},
			wantStatus: http.StatusNotFound,
			wantBody:   "Goal not found",
		},
		{
			name:       "ServiceError",
			id:         "goal-123",
			body:       `{"amount":100}`,
			mock:       &mocks.MockService{AddGoalProgressErr: apperrors.NewDatabase("db error", nil)},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			id:         "goal-123",
			body:       `{"amount":250}`,
			wantStatus: http.StatusOK,
			wantBody:   "Progress added",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/goals/"+tt.id+"/progress", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = []gin.Param{{Key: "id", Value: tt.id}}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.AddGoalProgress(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
