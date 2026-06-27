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

func TestDeleteGoal(t *testing.T) {
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
			name:       "NotFound",
			id:         "nonexistent",
			mock:       &mocks.MockService{DeleteGoalErr: apperrors.NewNotFound("Goal not found", nil)},
			wantStatus: http.StatusNotFound,
			wantBody:   "Goal not found",
		},
		{
			name:       "ServiceError",
			id:         "goal-123",
			mock:       &mocks.MockService{DeleteGoalErr: apperrors.NewDatabase("db error", nil)},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			id:         "goal-123",
			wantStatus: http.StatusOK,
			wantBody:   "Goal deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/goals/"+tt.id, nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.id}}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.DeleteGoal(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
