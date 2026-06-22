package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/mocks"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

func TestGetCategories(t *testing.T) {
	t.Parallel()

	cats := []models.Category{
		{ID: "id1", Name: "food", Icon: "fork", Color: "#FF0000"},
		{ID: "id2", Name: "transport", Icon: "car", Color: "#00FF00"},
	}

	tests := []struct {
		name       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name: "service error returns 500",
			mock: &mocks.MockService{
				GetCategoriesErr: apperrors.NewDatabase("failed to fetch categories", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success with empty list",
			mock: &mocks.MockService{
				GetCategoriesResult: []models.Category{},
			},
			wantStatus: http.StatusOK,
			wantBody:   "categories",
		},
		{
			name: "success with data",
			mock: &mocks.MockService{
				GetCategoriesResult: cats,
			},
			wantStatus: http.StatusOK,
			wantBody:   "id1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/categories", nil)

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.GetCategories(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
