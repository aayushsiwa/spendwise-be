package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

func TestHandler_RecalculateBalances(t *testing.T) {
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
			h.RecalculateBalances(tt.args.c)
		})
	}
}

func TestRecalculateBalances_ServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/refresh-balance", nil)

	svc := &mockService{
		refreshBalancesFn: func(_ context.Context) error {
			return fmt.Errorf("database failure")
		},
	}
	h := &Handler{Service: svc}
	h.RecalculateBalances(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "Internal Server Error" {
		t.Errorf("expected 'Internal Server Error', got %v", resp["error"])
	}
}

func TestRecalculateBalances_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/refresh-balance", nil)

	svc := &mockService{
		refreshBalancesFn: func(_ context.Context) error {
			return nil
		},
	}
	h := &Handler{Service: svc}
	h.RecalculateBalances(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "Balances recalculated successfully" {
		t.Errorf("expected success status message, got %v", resp["status"])
	}
}
