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

func TestUpdateGoal(t *testing.T) {
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
			body:       `{"name":"Save"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "InvalidJSON",
			id:         "goal-123",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_EmptyName",
			id:         "goal-123",
			body:       `{"name":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_NegativeTarget",
			id:         "goal-123",
			body:       `{"targetAmount":-100}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "NotFound",
			id:         "nonexistent",
			body:       `{"name":"Save"}`,
			mock:       &mocks.MockService{UpdateGoalErr: apperrors.NewNotFound("Goal not found", nil)},
			wantStatus: http.StatusNotFound,
			wantBody:   "Goal not found",
		},
		{
			name:       "ServiceError",
			id:         "goal-123",
			body:       `{"name":"Save"}`,
			mock:       &mocks.MockService{UpdateGoalErr: apperrors.NewDatabase("db error", nil)},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success_Name",
			id:         "goal-123",
			body:       `{"name":"New name"}`,
			wantStatus: http.StatusOK,
			wantBody:   "Goal updated",
		},
		{
			name:       "Success_TargetAmount",
			id:         "goal-123",
			body:       `{"targetAmount":5000}`,
			wantStatus: http.StatusOK,
			wantBody:   "Goal updated",
		},
		{
			name:       "Success_Status",
			id:         "goal-123",
			body:       `{"status":"achieved"}`,
			wantStatus: http.StatusOK,
			wantBody:   "Goal updated",
		},
		{
			name:       "Success_InvalidStatus",
			id:         "goal-123",
			body:       `{"status":"invalid"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPatch, "/goals/"+tt.id, strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = []gin.Param{{Key: "id", Value: tt.id}}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.UpdateGoal(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
