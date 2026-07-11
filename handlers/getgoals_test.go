package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/mocks"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

func TestGetGoals(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "ServiceError",
			mock:       &mocks.MockService{GetGoalsErr: errors.New("db error")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			mock:       &mocks.MockService{GetGoalsResult: []models.Goal{{ID: "1", Name: "Save"}}},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"Save"`,
		},
		{
			name:       "Empty",
			mock:       &mocks.MockService{GetGoalsResult: []models.Goal{}},
			wantStatus: http.StatusOK,
			wantBody:   `"goals":[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/goals", nil)

			h := &Handler{Service: tt.mock}
			h.GetGoals(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body = %s, want substring %s", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestGetGoal(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "InvalidID",
			id:         "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "NotFound",
			id:   "nonexistent",
			mock: &mocks.MockService{GetGoalFn: func(ctx context.Context, id string) (*models.Goal, error) {
				return nil, apperrors.NewNotFound("Goal not found", nil)
			}},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ServiceError",
			id:         "someid",
			mock:       &mocks.MockService{GetGoalErr: errors.New("db error")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			id:         "validid",
			mock:       &mocks.MockService{GetGoalResult: &models.Goal{ID: "validid", Name: "Vacation"}},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"Vacation"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "id", Value: tt.id}}
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/goals/"+tt.id, nil)

			ms := tt.mock
			if ms == nil {
				ms = &mocks.MockService{}
			}
			h := &Handler{Service: ms}
			h.GetGoal(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body = %s, want substring %s", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestGetGoal_RecordNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "someid"}}
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/goals/someid", nil)

	mock := &mocks.MockService{GetGoalFn: func(ctx context.Context, id string) (*models.Goal, error) {
		return nil, apperrors.NewNotFound("Goal not found", nil)
	}}
	h := &Handler{Service: mock}
	h.GetGoal(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
