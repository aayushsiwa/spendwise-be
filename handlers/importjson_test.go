package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aayushsiwa/expense-tracker/mocks"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

func TestImportJSON(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "invalid JSON",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
			wantBody:   "Invalid JSON array",
		},
		{
			name:       "not an array",
			body:       `{"date":"2024-01-01"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: func() string {
				records := []models.Record{
					{Date: "2024-01-01", Description: "Test", Category: "food", Amount: 50.0, Type: "expense"},
				}
				b, _ := json.Marshal(records)
				return string(b)
			}(),
			mock: &mocks.MockService{
				ImportJSONErr: fmt.Errorf("import failed"),
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal_error",
		},
		{
			name:       "success empty",
			body:       `[]`,
			wantStatus: http.StatusCreated,
			wantBody:   `"recordsImported":0`,
		},
		{
			name: "success",
			body: func() string {
				records := []models.Record{
					{Date: "2024-01-01", Description: "Salary", Category: "income", Amount: 3000.0, Type: "income"},
					{Date: "2024-01-02", Description: "Rent", Category: "housing", Amount: 1200.0, Type: "expense"},
				}
				b, _ := json.Marshal(records)
				return string(b)
			}(),
			wantStatus: http.StatusCreated,
			wantBody:   `"recordsImported":2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/import/json", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.ImportJSON(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
