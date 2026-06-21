package handlers

import (
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

func TestHandler_GetRecord(t *testing.T) {
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
			h.GetRecord(tt.args.c)
		})
	}
}

func TestGetRecord_EmptyID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records/", nil)
	c.Params = gin.Params{{Key: "id", Value: ""}}

	h := &Handler{Service: &mockService{}}
	h.GetRecord(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty ID, got %d", w.Code)
	}
}

func TestGetRecord_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records/nonexistent", nil)
	c.Params = gin.Params{{Key: "id", Value: "nonexistent"}}

	svc := &mockService{
		getRecordFn: func(_ context.Context, id string) (*models.Record, error) {
			return nil, apperrors.NewNotFound("Record not found", nil)
		},
	}
	h := &Handler{Service: svc}
	h.GetRecord(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetRecord_DatabaseError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records/someid", nil)
	c.Params = gin.Params{{Key: "id", Value: "someid"}}

	svc := &mockService{
		getRecordFn: func(_ context.Context, id string) (*models.Record, error) {
			return nil, apperrors.NewDatabase("database error", nil)
		},
	}
	h := &Handler{Service: svc}
	h.GetRecord(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetRecord_Success(t *testing.T) {
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

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records/"+recordID, nil)
	c.Params = gin.Params{{Key: "id", Value: recordID}}

	svc := &mockService{
		getRecordFn: func(_ context.Context, id string) (*models.Record, error) {
			return expected, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetRecord(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp models.Record
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response: %v", err)
	}
	if resp.ID != recordID {
		t.Errorf("expected ID %q, got %q", recordID, resp.ID)
	}
	if resp.Amount != 75.5 {
		t.Errorf("expected amount 75.5, got %f", resp.Amount)
	}
}
