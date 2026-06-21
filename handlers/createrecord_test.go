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

func TestCreateRecord(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mock       *mocks.MockService
		genIDFunc  func(string) (string, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:       "InvalidJSON",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_MissingDate",
			body:       `{"category":"food","amount":100,"type":"expense"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_MissingCategory",
			body:       `{"date":"2024-01-15","amount":100,"type":"expense"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_NegativeAmount",
			body:       `{"date":"2024-01-15","category":"food","amount":-50,"type":"expense"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_ZeroAmount",
			body:       `{"date":"2024-01-15","category":"food","amount":0,"type":"expense"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "ServiceError",
			body: `{"date":"2024-01-15","category":"food","amount":100,"type":"expense"}`,
			mock: &mocks.MockService{
				CreateRecordErr: apperrors.NewInvalidInput("Category not found", nil),
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "GenIDError",
			body: `{"date":"2024-01-15","category":"food","amount":100,"type":"expense"}`,
			genIDFunc: func(d string) (string, error) {
				return "", apperrors.NewInternal("Failed to generate record ID", nil)
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			body:       `{"date":"2024-01-15","category":"food","amount":100,"type":"expense"}`,
			wantStatus: http.StatusCreated,
			wantBody:   `"ID"`,
		},
		{
			name:       "ResponseMessageContainsID",
			body:       `{"date":"2024-01-15","category":"food","amount":25.5,"type":"income"}`,
			wantStatus: http.StatusCreated,
			wantBody:   `"ID"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/records", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			if tt.genIDFunc != nil {
				origGenIDFunc := genIDFunc
				genIDFunc = tt.genIDFunc
				defer func() { genIDFunc = origGenIDFunc }()
			}
			h := &Handler{
				Service: svc,
			}
			h.CreateRecord(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
