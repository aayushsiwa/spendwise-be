package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

func TestHandler_GetRecords(t *testing.T) {
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
			h.GetRecords(tt.args.c)
		})
	}
}

func Test_buildWhereClause(t *testing.T) {
	type args struct {
		q *models.QueryParams
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []any
	}{
		{
			name:  "no filters returns empty clause",
			args:  args{q: &models.QueryParams{}},
			want:  "",
			want1: []any{},
		},
		{
			name:  "type filter",
			args:  args{q: &models.QueryParams{Type: "income"}},
			want:  " WHERE r.type = ?",
			want1: []any{models.RecordType("income")},
		},
		{
			name:  "category filter",
			args:  args{q: &models.QueryParams{Category: "food"}},
			want:  " WHERE c.name = ?",
			want1: []any{"food"},
		},
		{
			name: "from and to date range",
			args: args{q: &models.QueryParams{
				PaginationFilterParams: models.PaginationFilterParams{},
				From:                   "2024-01-01",
				To:                     "2024-12-31",
			}},
			want:  " WHERE r.date >= ? AND r.date <= ?",
			want1: []any{"2024-01-01", "2024-12-31"},
		},
		{
			name:  "min and max amount",
			args:  args{q: &models.QueryParams{MinAmount: 10.0, MaxAmount: 100.0}},
			want:  " WHERE r.amount >= ? AND r.amount <= ?",
			want1: []any{10.0, 100.0},
		},
		{
			name:  "search term is lowercased and wrapped in wildcards",
			args:  args{q: &models.QueryParams{Search: "Grocery"}},
			want:  " WHERE LOWER(r.description) LIKE ?",
			want1: []any{"%grocery%"},
		},
		{
			name: "multiple filters combined",
			args: args{q: &models.QueryParams{
				Type:     "expense",
				Category: "food",
			}},
			want:  " WHERE r.type = ? AND c.name = ?",
			want1: []any{models.RecordType("expense"), "food"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := buildWhereClause(tt.args.q)
			if got != tt.want {
				t.Errorf("buildWhereClause() got = %q, want %q", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("buildWhereClause() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetRecords_ServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records", nil)

	svc := &mockService{
		getRecordsFn: func(_ context.Context, _ string, _ []any, _, _ int) ([]models.Record, int, error) {
			return nil, 0, apperrors.NewDatabase("failed to get records", nil)
		},
	}
	h := &Handler{Service: svc}
	h.GetRecords(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetRecords_SuccessWithDefaultPagination(t *testing.T) {
	recs := []models.Record{
		{ID: "r1", Amount: 50.0, Type: "expense"},
		{ID: "r2", Amount: 100.0, Type: "income"},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records", nil)

	svc := &mockService{
		getRecordsFn: func(_ context.Context, _ string, _ []any, _, _ int) ([]models.Record, int, error) {
			return recs, 2, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetRecords(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp models.RecordsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse response: %v", err)
	}
	if len(resp.Records) != 2 {
		t.Errorf("expected 2 records, got %d", len(resp.Records))
	}
	if resp.TotalCount != 2 {
		t.Errorf("expected total count 2, got %d", resp.TotalCount)
	}
}

func TestGetRecords_PaginationMetadata(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records?page=2&limit=5", nil)

	svc := &mockService{
		getRecordsFn: func(_ context.Context, _ string, _ []any, _, _ int) ([]models.Record, int, error) {
			return []models.Record{}, 12, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetRecords(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp models.RecordsResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Page != 2 {
		t.Errorf("expected page 2, got %d", resp.Page)
	}
	if resp.Limit != 5 {
		t.Errorf("expected limit 5, got %d", resp.Limit)
	}
	if resp.TotalPages != 3 {
		t.Errorf("expected 3 total pages (ceil(12/5)), got %d", resp.TotalPages)
	}
	if !resp.HasPrev {
		t.Error("expected hasPrev=true for page 2")
	}
	if !resp.HasNext {
		t.Error("expected hasNext=true (page 2 of 3)")
	}
}

func TestGetRecords_GroupedByCategory(t *testing.T) {
	groups := []models.GroupedRecord{
		{Group: "food", Total: 300.0, Count: 5},
		{Group: "transport", Total: 150.0, Count: 3},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records?groupBy=category", nil)

	svc := &mockService{
		getGroupedRecordsFn: func(_ context.Context, groupBy, _ string, _ []any) ([]models.GroupedRecord, error) {
			if groupBy != "category" {
				t.Errorf("expected groupBy=category, got %s", groupBy)
			}
			return groups, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetRecords(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp models.GroupedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not parse grouped response: %v", err)
	}
	if len(resp.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(resp.Groups))
	}
}

func TestGetRecords_GroupedServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records?groupBy=month", nil)

	svc := &mockService{
		getGroupedRecordsFn: func(_ context.Context, _, _ string, _ []any) ([]models.GroupedRecord, error) {
			return nil, apperrors.NewDatabase("grouped query failed", nil)
		},
	}
	h := &Handler{Service: svc}
	h.GetRecords(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestGetRecords_SearchFilter(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/records?search=coffee", nil)

	var capturedWhere string
	svc := &mockService{
		getRecordsFn: func(_ context.Context, whereClause string, _ []any, _, _ int) ([]models.Record, int, error) {
			capturedWhere = whereClause
			return []models.Record{}, 0, nil
		},
	}
	h := &Handler{Service: svc}
	h.GetRecords(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(capturedWhere, "LOWER(r.description) LIKE") {
		t.Errorf("expected WHERE clause with LIKE, got %q", capturedWhere)
	}
}
