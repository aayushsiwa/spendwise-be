package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"

	_ "modernc.org/sqlite"
)

func TestHandler_ExportCSV(t *testing.T) {
	type fields struct {
		Service services.Service
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				Service: tt.fields.Service,
			}
			h.ExportCSV(tt.args.c)
		})
	}
}

func TestExportCSV_ServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/export/csv", nil)

	svc := &mockService{
		exportRecordsFn: func(_ context.Context) (*sql.Rows, error) {
			return nil, errors.New("database error")
		},
	}
	h := &Handler{Service: svc}
	h.ExportCSV(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Error querying records") {
		t.Errorf("expected error message in body, got: %s", w.Body.String())
	}
}

func TestExportCSV_InlineContentType(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT '2024-01-15', 'groceries', 'food', 100.0, 'expense', 'weekly shopping'",
	)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/export/csv", nil)

	svc := &mockService{
		exportRecordsFn: func(_ context.Context) (*sql.Rows, error) {
			return rows, nil
		},
	}
	h := &Handler{Service: svc}
	h.ExportCSV(c)

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain content-type for inline, got %q", ct)
	}
	cd := w.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "inline") {
		t.Errorf("expected inline content-disposition, got %q", cd)
	}
}

func TestExportCSV_DownloadContentType(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT '2024-01-15', 'groceries', 'food', 100.0, 'expense', 'note'",
	)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/export/csv?download=true", nil)

	svc := &mockService{
		exportRecordsFn: func(_ context.Context) (*sql.Rows, error) {
			return rows, nil
		},
	}
	h := &Handler{Service: svc}
	h.ExportCSV(c)

	ct := w.Header().Get("Content-Type")
	if ct != "text/csv" {
		t.Errorf("expected text/csv content-type for download, got %q", ct)
	}
	cd := w.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "attachment") {
		t.Errorf("expected attachment content-disposition, got %q", cd)
	}
}

func TestExportCSV_CSVHeaderAlwaysPresent(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	defer db.Close()
	rows, _ := db.Query("SELECT '2024-01-01', 'desc', 'cat', 10.0, 'income', 'note'")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/export/csv?download=true", nil)

	svc := &mockService{
		exportRecordsFn: func(_ context.Context) (*sql.Rows, error) {
			return rows, nil
		},
	}
	h := &Handler{Service: svc}
	h.ExportCSV(c)

	body := w.Body.String()
	if !strings.Contains(body, "Date") || !strings.Contains(body, "Description") {
		t.Errorf("expected CSV header row in body, got: %s", body)
	}
}
