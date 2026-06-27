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

func TestCreateGoal(t *testing.T) {
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
			name:       "ValidationError_MissingName",
			body:       `{"targetAmount":1000}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_MissingTargetAmount",
			body:       `{"name":"Save for vacation"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_NegativeTargetAmount",
			body:       `{"name":"Save","targetAmount":-100}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ServiceError",
			body:       `{"name":"Save","targetAmount":1000}`,
			mock:       &mocks.MockService{CreateGoalErr: apperrors.NewDatabase("db error", nil)},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			body:       `{"name":"Save","targetAmount":1000,"description":"Vacation fund"}`,
			mock:       &mocks.MockService{},
			wantStatus: http.StatusCreated,
			wantBody:   `"message":"Goal created"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/goals", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			h := &Handler{Service: tt.mock}
			if tt.mock == nil {
				h.Service = &mocks.MockService{}
			}
			h.CreateGoal(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("body = %s, want substring %s", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestCreateGoal_ServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"name":"test","targetAmount":500}`
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/goals", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mock := &mocks.MockService{}
	h := &Handler{Service: mock}
	h.CreateGoal(c)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}
