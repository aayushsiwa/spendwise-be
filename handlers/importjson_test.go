package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

func TestHandler_ImportJSON(t *testing.T) {
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
			h.ImportJSON(tt.args.c)
		})
	}
}

func TestImportJSON_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/import/json", bytes.NewBufferString("not-json"))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.ImportJSON(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "Invalid JSON array" {
		t.Errorf("expected 'Invalid JSON array' error, got %v", resp["error"])
	}
}

func TestImportJSON_NotAnArray(t *testing.T) {
	// A single object instead of array should fail ShouldBindJSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/import/json",
		bytes.NewBufferString(`{"date":"2024-01-01"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.ImportJSON(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-array JSON, got %d", w.Code)
	}
}

func TestImportJSON_ServiceError(t *testing.T) {
	records := []models.Record{
		{Date: "2024-01-01", Description: "Test", Category: "food", Amount: 50.0, Type: "expense"},
	}
	body, _ := json.Marshal(records)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/import/json", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	svc := &mockService{
		importJSONFn: func(_ context.Context, _ []models.Record) (int, error) {
			return 0, fmt.Errorf("import failed")
		},
	}
	h := &Handler{Service: svc}
	h.ImportJSON(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "JSON import failed" {
		t.Errorf("expected 'JSON import failed' error, got %v", resp["error"])
	}
}

func TestImportJSON_SuccessEmpty(t *testing.T) {
	body := bytes.NewBufferString(`[]`)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/import/json", body)
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.ImportJSON(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 for empty array, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["recordsImported"] != float64(0) {
		t.Errorf("expected recordsImported=0, got %v", resp["recordsImported"])
	}
}

func TestImportJSON_Success(t *testing.T) {
	records := []models.Record{
		{Date: "2024-01-01", Description: "Salary", Category: "income", Amount: 3000.0, Type: "income"},
		{Date: "2024-01-02", Description: "Rent", Category: "housing", Amount: 1200.0, Type: "expense"},
	}
	body, _ := json.Marshal(records)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/import/json", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	svc := &mockService{
		importJSONFn: func(_ context.Context, recs []models.Record) (int, error) {
			return len(recs), nil
		},
	}
	h := &Handler{Service: svc}
	h.ImportJSON(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["recordsImported"] != float64(2) {
		t.Errorf("expected recordsImported=2, got %v", resp["recordsImported"])
	}
	msg, _ := resp["message"].(string)
	if msg == "" {
		t.Error("expected non-empty message in response")
	}
}
