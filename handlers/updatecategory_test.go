package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/mocks"

	"github.com/gin-gonic/gin"
)

func TestUpdateCategory(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "EmptyID",
			id:         "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "InvalidJSON",
			id:         "cat-123",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_MissingName",
			id:         "cat-123",
			body:       `{"name":"","color":"#FF0000"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_InvalidColor",
			id:         "cat-123",
			body:       `{"name":"food","color":"not-a-color"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_InvalidNameCharacters",
			id:         "cat-123",
			body:       `{"name":"food@#$","color":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "ServiceNotFound",
			id:   "cat-999",
			body: `{"name":"food","color":"#FF0000"}`,
			mock: &mocks.MockService{
				UpdateCategoryErr: apperrors.NewNotFound("Category not found", nil),
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "ServiceDatabaseError",
			id:   "cat-999",
			body: `{"name":"food","color":"#FF0000"}`,
			mock: &mocks.MockService{
				UpdateCategoryErr: apperrors.NewDatabase("failed to update category", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			id:         "cat-abc-123",
			body:       `{"name":"food","icon":"star","color":"#AABBCC"}`,
			wantStatus: http.StatusOK,
			wantBody:   "cat-abc-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/categories/"+tt.id, strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.UpdateCategory(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
