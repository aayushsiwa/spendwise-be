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

func TestPatchRecord(t *testing.T) {
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
			id:         "rec-123",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_InvalidDate",
			id:         "rec-123",
			body:       `{"date":"bad-date"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ValidationError_NegativeAmount",
			id:         "rec-123",
			body:       `{"amount":-5}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "ServiceNotFound",
			id:   "rec-999",
			body: `{"amount":50}`,
			mock: &mocks.MockService{
				PatchRecordErr: apperrors.NewNotFound("Record not found", nil),
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "ServiceDatabaseError",
			id:   "rec-999",
			body: `{"amount":50}`,
			mock: &mocks.MockService{
				PatchRecordErr: apperrors.NewDatabase("failed to update record", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			id:         "rec-abc-123",
			body:       `{"amount":75.5}`,
			wantStatus: http.StatusOK,
			wantBody:   "Record updated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/records/"+tt.id, strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.PatchRecord(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
