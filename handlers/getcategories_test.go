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

func TestHandler_GetCategories(t *testing.T) {
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
			h.GetCategories(tt.args.c)
		})
	}
}

func TestGetCategories_ServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/categories", nil)

	svc := &mockService{
		getCategoriesFn: func(_ context.Context) ([]models.Category, error) {
			return nil, apperrors.NewDatabase("failed to fetch categories", nil)
		},
	}
	h := &Handler{Service: svc}
	h.GetCategories(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetCategories_SuccessEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/categories", nil)

	svc := &mockService{
		getCategoriesFn: func(_ context.Context) ([]models.Category, error) {
			return []models.Category{}, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetCategories(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response: %v", err)
	}
	if _, ok := resp["categories"]; !ok {
		t.Error("expected 'categories' key in response")
	}
}

func TestGetCategories_SuccessWithData(t *testing.T) {
	cats := []models.Category{
		{ID: "id1", Name: "food", Icon: "fork", Color: "#FF0000"},
		{ID: "id2", Name: "transport", Icon: "car", Color: "#00FF00"},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/categories", nil)

	svc := &mockService{
		getCategoriesFn: func(_ context.Context) ([]models.Category, error) {
			return cats, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetCategories(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	catList, ok := resp["categories"].([]any)
	if !ok {
		t.Fatal("expected 'categories' to be an array")
	}
	if len(catList) != 2 {
		t.Errorf("expected 2 categories, got %d", len(catList))
	}
}
