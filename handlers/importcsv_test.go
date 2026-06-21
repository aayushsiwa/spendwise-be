package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

func TestHandler_ImportCSV(t *testing.T) {
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
			h.ImportCSV(tt.args.c)
		})
	}
}

// buildCSVUploadRequest creates a multipart form request with a CSV file.
func buildCSVUploadRequest(t *testing.T, content string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "records.csv")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := io.WriteString(part, content); err != nil {
		t.Fatalf("failed to write csv content: %v", err)
	}
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/import/csv", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestImportCSV_NoFileProvided(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/import/csv", nil)
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.ImportCSV(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when no file provided, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "CSV file not provided" {
		t.Errorf("expected 'CSV file not provided' error, got %v", resp["error"])
	}
}

func TestImportCSV_FileTooLarge(t *testing.T) {
	// Build a multipart request with a "file" field that is just over 10 MB
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.csv")
	// Write 10MB + 1 byte
	chunk := strings.Repeat("a", 1024)
	for i := 0; i < 10*1024+1; i++ {
		_, _ = io.WriteString(part, chunk)
	}
	writer.Close()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/import/csv", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	h := &Handler{Service: &mockService{}}
	h.ImportCSV(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for oversized file, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "File too large" {
		t.Errorf("expected 'File too large' error, got %v", resp["error"])
	}
}

func TestImportCSV_ServiceError(t *testing.T) {
	csvContent := "date,description,amount\n2024-01-01,Test,100\n"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = buildCSVUploadRequest(t, csvContent)

	svc := &mockService{
		importCSVFn: func(_ context.Context, _ io.Reader) (int, int, error) {
			return 0, 0, fmt.Errorf("service error")
		},
	}
	h := &Handler{Service: svc}
	h.ImportCSV(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestImportCSV_Success(t *testing.T) {
	csvContent := "date,description,amount\n2024-01-01,Groceries,50\n2024-01-02,Coffee,5\n"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = buildCSVUploadRequest(t, csvContent)

	svc := &mockService{
		importCSVFn: func(_ context.Context, _ io.Reader) (int, int, error) {
			return 2, 0, nil
		},
	}
	h := &Handler{Service: svc}
	h.ImportCSV(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["recordsImported"] != float64(2) {
		t.Errorf("expected recordsImported=2, got %v", resp["recordsImported"])
	}
	if resp["skippedCount"] != float64(0) {
		t.Errorf("expected skippedCount=0, got %v", resp["skippedCount"])
	}
}

func TestImportCSV_SuccessWithSkipped(t *testing.T) {
	csvContent := "date,description,amount\n2024-01-01,Good,50\n,Bad,\n"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = buildCSVUploadRequest(t, csvContent)

	svc := &mockService{
		importCSVFn: func(_ context.Context, _ io.Reader) (int, int, error) {
			return 1, 1, nil
		},
	}
	h := &Handler{Service: svc}
	h.ImportCSV(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["skippedCount"] != float64(1) {
		t.Errorf("expected skippedCount=1, got %v", resp["skippedCount"])
	}
}
