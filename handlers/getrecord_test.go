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

func TestGetRecord(t *testing.T) {
	t.Parallel()

	recordID := "rec-001"
	expected := &models.Record{
		ID:          recordID,
		Date:        "2024-01-15",
		Description: "Grocery shopping",
		Category:    "food",
		Amount:      75.5,
		Type:        "expense",
		Note:        "weekly",
		Balance:     924.5,
	}

	tests := []struct {
		name       string
		params     gin.Params
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "empty ID returns 400",
			params:     gin.Params{{Key: "id", Value: ""}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "not found returns 404",
			params: gin.Params{{Key: "id", Value: "nonexistent"}},
			mock: &mocks.MockService{
				GetRecordErr: apperrors.NewNotFound("Record not found", nil),
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "database error returns 500",
			params: gin.Params{{Key: "id", Value: "someid"}},
			mock: &mocks.MockService{
				GetRecordErr: apperrors.NewDatabase("database error", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:   "success returns record",
			params: gin.Params{{Key: "id", Value: recordID}},
			mock: &mocks.MockService{
				GetRecordResult: expected,
			},
			wantStatus: http.StatusOK,
			wantBody:   recordID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			id := tt.params.ByName("id")
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/records/"+id, nil)
			c.Params = tt.params

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.GetRecord(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
