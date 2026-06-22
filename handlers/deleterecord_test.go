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

func TestDeleteRecord(t *testing.T) {
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
			id:   "xyz789",
			mock: &mocks.MockService{
				DeleteRecordErr: apperrors.NewNotFound("Record not found", nil),
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "DatabaseError",
			id:   "xyz789",
			mock: &mocks.MockService{
				DeleteRecordErr: apperrors.NewDatabase("Failed to delete record", nil),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Success",
			id:         "rec-abc-123",
			wantStatus: http.StatusOK,
			wantBody:   "rec-abc-123",
		},
		{
			name:       "ResponseMessageFormat",
			id:         "some-id-456",
			wantStatus: http.StatusOK,
			wantBody:   "some-id-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/records/"+tt.id, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.DeleteRecord(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
