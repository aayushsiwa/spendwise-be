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

func TestHandler_CreateCategories(t *testing.T) {
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
			h.CreateCategories(tt.args.c)
		})
	}
}

func TestCreateCategories_InvalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString("not-json"))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateCategories(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateCategories_EmptyArray(t *testing.T) {
	body := bytes.NewBufferString(`[]`)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/categories", body)
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateCategories(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateCategories_ValidationError_MissingName(t *testing.T) {
	cats := []models.Category{{Name: "", Icon: "star", Color: "#FF0000"}}
	body, _ := json.Marshal(cats)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateCategories(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing category name, got %d", w.Code)
	}
}

func TestCreateCategories_ValidationError_InvalidColor(t *testing.T) {
	cats := []models.Category{{Name: "food", Icon: "star", Color: "not-a-color"}}
	body, _ := json.Marshal(cats)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateCategories(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid color, got %d", w.Code)
	}
}

func TestCreateCategories_ValidationError_InvalidCharactersInName(t *testing.T) {
	cats := []models.Category{{Name: "food@#$", Icon: "", Color: ""}}
	body, _ := json.Marshal(cats)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateCategories(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid name characters, got %d", w.Code)
	}
}

func TestCreateCategories_ServiceError(t *testing.T) {
	cats := []models.Category{{Name: "food", Icon: "", Color: ""}}
	body, _ := json.Marshal(cats)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	svc := &mockService{
		createCategoriesFn: func(_ context.Context, _ []models.Category) ([]models.Category, error) {
			return nil, apperrors.NewDatabase("failed to insert", nil)
		},
	}
	h := &Handler{Service: svc}
	h.CreateCategories(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestCreateCategories_Success(t *testing.T) {
	cats := []models.Category{{Name: "food", Icon: "fork", Color: "#AABBCC"}}
	body, _ := json.Marshal(cats)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	returned := []models.Category{{ID: "abc123", Name: "food", Icon: "fork", Color: "#AABBCC"}}
	svc := &mockService{
		createCategoriesFn: func(_ context.Context, _ []models.Category) ([]models.Category, error) {
			return returned, nil
		},
	}
	h := &Handler{Service: svc}
	h.CreateCategories(c)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 category in response, got %d", len(resp))
	}
	if resp[0]["ID"] != "abc123" {
		t.Errorf("expected ID abc123, got %v", resp[0]["ID"])
	}
}

func TestCreateCategories_IndexedValidationFieldNames(t *testing.T) {
	// Validation errors should include index: categories[1].name
	cats := []models.Category{
		{Name: "food", Icon: "", Color: ""},
		{Name: "", Icon: "", Color: ""},
	}
	body, _ := json.Marshal(cats)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{Service: &mockService{}}
	h.CreateCategories(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	errBlock, ok := resp["error"].(map[string]any)
	if !ok {
		t.Fatal("expected 'error' key in response")
	}
	details, ok := errBlock["details"].(map[string]any)
	if !ok {
		t.Fatal("expected 'details' key in error")
	}
	// The field should be indexed as categories[1].name
	if _, found := details["categories[1].name"]; !found {
		t.Errorf("expected field key 'categories[1].name' in details, got %v", details)
	}
}
