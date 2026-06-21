package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

func TestHandler_CreateRecord(t *testing.T) {
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
			h.CreateRecord(tt.args.c)
		})
	}
}

func TestCreateRecord_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/records", bytes.NewBufferString("not-json"))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateRecord(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateRecord_ValidationError_MissingDate(t *testing.T) {
	rec := models.Record{
		Category: "food",
		Amount:   100.0,
		Type:     "expense",
	}
	body, _ := json.Marshal(rec)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/records", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateRecord(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing date, got %d", w.Code)
	}
}

func TestCreateRecord_ValidationError_MissingCategory(t *testing.T) {
	rec := models.Record{
		Date:   "2024-01-15",
		Amount: 100.0,
		Type:   "expense",
	}
	body, _ := json.Marshal(rec)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/records", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateRecord(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing category, got %d", w.Code)
	}
}

func TestCreateRecord_ValidationError_NegativeAmount(t *testing.T) {
	rec := models.Record{
		Date:     "2024-01-15",
		Category: "food",
		Amount:   -50.0,
		Type:     "expense",
	}
	body, _ := json.Marshal(rec)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/records", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateRecord(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative amount, got %d", w.Code)
	}
}

func TestCreateRecord_ValidationError_ZeroAmount(t *testing.T) {
	rec := models.Record{
		Date:     "2024-01-15",
		Category: "food",
		Amount:   0,
		Type:     "expense",
	}
	body, _ := json.Marshal(rec)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/records", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateRecord(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for zero amount, got %d", w.Code)
	}
}

func TestCreateRecord_ServiceError(t *testing.T) {
	rec := models.Record{
		Date:     "2024-01-15",
		Category: "food",
		Amount:   100.0,
		Type:     "expense",
	}
	body, _ := json.Marshal(rec)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/records", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	svc := &mockService{
		createRecordFn: func(_ context.Context, _ *models.Record) error {
			return apperrors.NewInvalidInput("Category not found", nil)
		},
	}
	h := &Handler{Service: svc}
	h.CreateRecord(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for category not found, got %d", w.Code)
	}
}

func TestCreateRecord_Success(t *testing.T) {
	rec := models.Record{
		Date:     "2024-01-15",
		Category: "food",
		Amount:   100.0,
		Type:     "expense",
	}
	body, _ := json.Marshal(rec)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/records", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	svc := &mockService{
		createRecordFn: func(_ context.Context, _ *models.Record) error {
			return nil
		},
	}
	h := &Handler{Service: svc}
	h.CreateRecord(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response: %v", err)
	}
	if _, ok := resp["ID"]; !ok {
		t.Error("expected 'ID' field in response")
	}
	if _, ok := resp["message"]; !ok {
		t.Error("expected 'message' field in response")
	}
}

func TestCreateRecord_ResponseMessageContainsID(t *testing.T) {
	rec := models.Record{
		Date:     "2024-01-15",
		Category: "food",
		Amount:   25.5,
		Type:     "income",
	}
	body, _ := json.Marshal(rec)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/records", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateRecord(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	id, ok := resp["ID"].(string)
	if !ok || id == "" {
		t.Error("expected non-empty ID in response")
	}
	msg, _ := resp["message"].(string)
	if msg == "" {
		t.Error("expected non-empty message in response")
	}
}
