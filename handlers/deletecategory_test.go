package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

func TestHandler_DeleteCategory(t *testing.T) {
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
			h.DeleteCategory(tt.args.c)
		})
	}
}

func TestDeleteCategory_EmptyID(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/categories/", nil)
	c.Params = gin.Params{{Key: "id", Value: ""}}

	h := &Handler{Service: &mockService{}}
	h.DeleteCategory(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty ID, got %d", w.Code)
	}
}

func TestDeleteCategory_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/categories/abc123", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc123"}}

	svc := &mockService{
		deleteCategoryFn: func(_ context.Context, id string) error {
			return apperrors.NewNotFound("Category not found", nil)
		},
	}
	h := &Handler{Service: svc}
	h.DeleteCategory(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDeleteCategory_Conflict(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/categories/abc123", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc123"}}

	svc := &mockService{
		deleteCategoryFn: func(_ context.Context, id string) error {
			return apperrors.NewConflict("Cannot delete category that has associated records", nil)
		},
	}
	h := &Handler{Service: svc}
	h.DeleteCategory(c)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestDeleteCategory_DatabaseError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/categories/abc123", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc123"}}

	svc := &mockService{
		deleteCategoryFn: func(_ context.Context, id string) error {
			return apperrors.NewDatabase("Failed to delete category", nil)
		},
	}
	h := &Handler{Service: svc}
	h.DeleteCategory(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestDeleteCategory_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/categories/abc123", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc123"}}

	svc := &mockService{
		deleteCategoryFn: func(_ context.Context, id string) error {
			return nil
		},
	}
	h := &Handler{Service: svc}
	h.DeleteCategory(c)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}
