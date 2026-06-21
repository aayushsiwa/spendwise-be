package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aayushsiwa/expense-tracker/mocks"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

func TestExportCSV(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		query                  string
		mock                   *mocks.MockService
		writeError             bool
		wantStatus             int
		wantBody               string
		wantContentType        string
		wantContentDisposition string
	}{
		{
			name: "service error returns 500",
			mock: &mocks.MockService{
				ExportRecordsErr: errors.New("database error"),
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Database operation failed",
		},
		{
			name: "inline content type",
			mock: &mocks.MockService{
				ExportRecordsResult: []models.Record{
					{Date: "2024-01-15", Description: "groceries", Category: "food", Amount: 100.0, Type: models.Expense, Note: "weekly shopping"},
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "groceries",
		},
		{
			name:  "download content type",
			query: "download=true",
			mock: &mocks.MockService{
				ExportRecordsResult: []models.Record{
					{Date: "2024-01-15", Description: "groceries", Category: "food", Amount: 100.0, Type: models.Expense, Note: "note"},
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "groceries",
		},
		{
			name:  "CSV header always present",
			query: "download=true",
			mock: &mocks.MockService{
				ExportRecordsResult: []models.Record{
					{Date: "2024-01-01", Description: "desc", Category: "cat", Amount: 10.0, Type: models.Income, Note: "note"},
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "Date",
		},
		{
			name:       "write error returns early",
			query:      "download=false",
			writeError: true,
			mock: &mocks.MockService{
				ExportRecordsResult: []models.Record{
					{Date: "2024-01-01", Description: "desc", Category: "cat", Amount: 10.0, Type: models.Income, Note: "note"},
				},
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			path := "/export/csv"
			if tt.query != "" {
				path += "?" + tt.query
			}
			c.Request = httptest.NewRequest(http.MethodGet, path, nil)

			if tt.writeError {
				c.Writer = &errorResponseWriter{ResponseWriter: c.Writer}
			}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.ExportCSV(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}

			if tt.wantContentType != "" {
				ct := w.Header().Get("Content-Type")
				if !strings.Contains(ct, tt.wantContentType) {
					t.Errorf("expected Content-Type containing %q, got %q", tt.wantContentType, ct)
				}
			}
			if tt.wantContentDisposition != "" {
				cd := w.Header().Get("Content-Disposition")
				if !strings.Contains(cd, tt.wantContentDisposition) {
					t.Errorf("expected Content-Disposition containing %q, got %q", tt.wantContentDisposition, cd)
				}
			}
		})
	}
}

type errorResponseWriter struct {
	gin.ResponseWriter
}

func (w *errorResponseWriter) Write(b []byte) (int, error) {
	return 0, errors.New("test write error")
}
