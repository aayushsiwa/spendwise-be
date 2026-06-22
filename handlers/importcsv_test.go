package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aayushsiwa/expense-tracker/mocks"
	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

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
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/import/csv", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func buildLargeCSVUploadRequest(t *testing.T) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.csv")
	chunk := strings.Repeat("a", 1024)
	for range 10*1024 + 1 {
		_, _ = io.WriteString(part, chunk)
	}
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/import/csv", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestImportCSV(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
		buildReq   func(t *testing.T) *http.Request
	}{
		{
			name:       "no file provided",
			wantStatus: http.StatusBadRequest,
			wantBody:   "CSV file not provided",
		},
		{
			name:       "file too large",
			mock:       &mocks.MockService{},
			wantStatus: http.StatusBadRequest,
			wantBody:   "File too large",
			buildReq:   buildLargeCSVUploadRequest,
		},
		{
			name: "import validation error returns 422",
			mock: &mocks.MockService{
				ImportCSVErr: services.ErrImportValidation,
			},
			wantStatus: http.StatusUnprocessableEntity,
			buildReq: func(t *testing.T) *http.Request {
				return buildCSVUploadRequest(t, "date,description,amount\n2024-01-01,Test,100\n")
			},
		},
		{
			name: "service error",
			mock: &mocks.MockService{
				ImportCSVErr: fmt.Errorf("service error"),
			},
			wantStatus: http.StatusInternalServerError,
			buildReq: func(t *testing.T) *http.Request {
				return buildCSVUploadRequest(t, "date,description,amount\n2024-01-01,Test,100\n")
			},
		},
		{
			name: "success",
			mock: &mocks.MockService{
				ImportCSVFn: func(_ context.Context, _ io.Reader) (int, int, error) {
					return 2, 0, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"recordsImported":2`,
			buildReq: func(t *testing.T) *http.Request {
				return buildCSVUploadRequest(t, "date,description,amount\n2024-01-01,Groceries,50\n2024-01-02,Coffee,5\n")
			},
		},
		{
			name: "success with skipped",
			mock: &mocks.MockService{
				ImportCSVFn: func(_ context.Context, _ io.Reader) (int, int, error) {
					return 1, 1, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"skippedCount":1`,
			buildReq: func(t *testing.T) *http.Request {
				return buildCSVUploadRequest(t, "date,description,amount\n2024-01-01,Good,50\n,Bad,\n")
			},
		},
		{
			name:       "open file error",
			mock:       &mocks.MockService{},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Failed to open uploaded file",
			buildReq: func(t *testing.T) *http.Request {
				return buildCSVUploadRequest(t, "date,description,amount\n2024-01-01,Good,50\n")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.buildReq != nil {
				c.Request = tt.buildReq(t)
			} else {
				c.Request = httptest.NewRequest(http.MethodPost, "/import/csv", strings.NewReader(tt.body))
				c.Request.Header.Set("Content-Type", "application/json")
			}

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			if tt.name == "open file error" {
				origOpenFileFunc := openFileFunc
				openFileFunc = func(fh *multipart.FileHeader) (multipart.File, error) {
					return nil, fmt.Errorf("open file error")
				}
				defer func() { openFileFunc = origOpenFileFunc }()
			}
			h := &Handler{Service: svc}
			h.ImportCSV(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
