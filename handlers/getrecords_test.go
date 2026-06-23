package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apperrors "aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/mocks"
	"aayushsiwa/expense-tracker/models"

	"github.com/gin-gonic/gin"
)

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
