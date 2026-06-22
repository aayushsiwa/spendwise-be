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

func TestCreateCategories(t *testing.T) {
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
			name:       "EmptyArray",
			body:       `[]`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_MissingName",
			body:       `[{"name":"","icon":"star","color":"#FF0000"}]`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_InvalidColor",
			body:       `[{"name":"food","icon":"star","color":"not-a-color"}]`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_InvalidCharactersInName",
			body:       `[{"name":"food@#$","icon":"","color":""}]`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "ServiceError",
			body: `[{"name":"food","icon":"","color":""}]`,
			mock: &mocks.MockService{
				CreateCategoriesErr: apperrors.NewDatabase("failed to insert", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "Success",
			body: `[{"name":"food","icon":"fork","color":"#AABBCC"}]`,
			mock: &mocks.MockService{
				CreateCategoriesResult: []models.Category{{ID: "abc123", Name: "food", Icon: "fork", Color: "#AABBCC"}},
			},
			wantStatus: http.StatusCreated,
			wantBody:   "abc123",
		},
		{
			name:       "IndexedValidationFieldNames",
			body:       `[{"name":"food","icon":"","color":""},{"name":"","icon":"","color":""}]`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "categories[1].name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/categories", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.CreateCategories(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
