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

func TestDeleteCategory(t *testing.T) {
	tests := []struct {
		name       string
		id         string
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
			name: "NotFound",
			id:   "abc123",
			mock: &mocks.MockService{
				DeleteCategoryErr: apperrors.NewNotFound("Category not found", nil),
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "Conflict",
			id:   "abc123",
			mock: &mocks.MockService{
				DeleteCategoryErr: apperrors.NewConflict("Cannot delete category that has associated records", nil),
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "DatabaseError",
			id:   "abc123",
			mock: &mocks.MockService{
				DeleteCategoryErr: apperrors.NewDatabase("Failed to delete category", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			id:         "abc123",
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/categories/"+tt.id, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.DeleteCategory(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
