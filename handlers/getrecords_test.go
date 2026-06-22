package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/mocks"
	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

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
			got, got1 := services.BuildWhereClause(tt.args.q)
			if got != tt.want {
				t.Errorf("buildWhereClause() got = %q, want %q", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("buildWhereClause() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetRecords(t *testing.T) {
	recs := []models.Record{
		{ID: "r1", Amount: 50.0, Type: "expense"},
		{ID: "r2", Amount: 100.0, Type: "income"},
	}
	groups := []models.GroupedRecord{
		{Group: "food", Total: 300.0, Count: 5},
		{Group: "transport", Total: 150.0, Count: 3},
	}

	tests := []struct {
		name       string
		query      string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name:       "service error returns 500",
			query:      "/records",
			mock:       &mocks.MockService{GetRecordsErr: apperrors.NewDatabase("failed to get records", nil)},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "success with default pagination",
			query:      "/records",
			mock:       &mocks.MockService{GetRecordsResult: recs, GetRecordsTotalCount: 2},
			wantStatus: http.StatusOK,
			wantBody:   `"records"`,
		},
		{
			name:       "pagination metadata",
			query:      "/records?page=2&limit=5",
			mock:       &mocks.MockService{GetRecordsResult: []models.Record{}, GetRecordsTotalCount: 12},
			wantStatus: http.StatusOK,
			wantBody:   `"page":2`,
		},
		{
			name:       "grouped by category",
			query:      "/records?groupBy=category",
			mock:       &mocks.MockService{GetGroupedRecordsResult: groups},
			wantStatus: http.StatusOK,
			wantBody:   `"groups"`,
		},
		{
			name:       "grouped service error returns 500",
			query:      "/records?groupBy=month",
			mock:       &mocks.MockService{GetGroupedRecordsErr: apperrors.NewDatabase("grouped query failed", nil)},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid query params returns 400",
			query:      "/records?page=-1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "grouped by month returns groups",
			query:      "/records?groupBy=month",
			mock:       &mocks.MockService{GetGroupedRecordsResult: groups},
			wantStatus: http.StatusOK,
			wantBody:   `"groups"`,
		},
		{
			name:  "search filter passes params to service",
			query: "/records?search=coffee",
			mock: &mocks.MockService{GetRecordsFn: func(_ context.Context, params *models.QueryParams) ([]models.Record, int, error) {
				if params.Search != "coffee" {
					t.Errorf("expected search=coffee, got %q", params.Search)
				}
				return []models.Record{}, 0, nil
			}},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.query, nil)

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.GetRecords(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
