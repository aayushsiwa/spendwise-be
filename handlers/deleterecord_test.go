package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"

	"github.com/gin-gonic/gin"
)

func TestDeleteRecord_EmptyID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/records/", nil)
	c.Params = gin.Params{{Key: "id", Value: ""}}

	h := &Handler{Service: &mockService{}}
	h.DeleteRecord(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty ID, got %d", w.Code)
	}
}

func TestDeleteRecord_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/records/xyz789", nil)
	c.Params = gin.Params{{Key: "id", Value: "xyz789"}}

	svc := &mockService{
		deleteRecordFn: func(_ context.Context, id string) (int64, error) {
			return 0, apperrors.NewNotFound("Record not found", nil)
		},
	}
	h := &Handler{Service: svc}
	h.DeleteRecord(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDeleteRecord_DatabaseError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/records/xyz789", nil)
	c.Params = gin.Params{{Key: "id", Value: "xyz789"}}

	svc := &mockService{
		deleteRecordFn: func(_ context.Context, id string) (int64, error) {
			return 0, apperrors.NewDatabase("Failed to delete record", nil)
		},
	}
	h := &Handler{Service: svc}
	h.DeleteRecord(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestDeleteRecord_Success(t *testing.T) {
	recordID := "rec-abc-123"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/records/"+recordID, nil)
	c.Params = gin.Params{{Key: "id", Value: recordID}}

	svc := &mockService{
		deleteRecordFn: func(_ context.Context, id string) (int64, error) {
			return 1, nil
		},
	}
	h := &Handler{Service: svc}
	h.DeleteRecord(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response: %v", err)
	}
	msg, _ := resp["message"].(string)
	if !strings.Contains(msg, recordID) {
		t.Errorf("expected message to contain record ID %q, got %q", recordID, msg)
	}
}

func TestDeleteRecord_ResponseMessageFormat(t *testing.T) {
	recordID := "some-id-456"
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/records/"+recordID, nil)
	c.Params = gin.Params{{Key: "id", Value: recordID}}

	h := &Handler{Service: &mockService{}}
	h.DeleteRecord(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	expected := fmt.Sprintf("Record with id %s deleted successfully", recordID)
	if resp["message"] != expected {
		t.Errorf("expected message %q, got %v", expected, resp["message"])
	}
}