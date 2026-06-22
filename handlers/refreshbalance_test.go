package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aayushsiwa/expense-tracker/mocks"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestRecalculateBalances(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mock       *mocks.MockService
		wantStatus int
		wantBody   string
	}{
		{
			name: "service error",
			mock: &mocks.MockService{
				RefreshBalancesErr: fmt.Errorf("database failure"),
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "database_error",
		},
		{
			name:       "success",
			wantStatus: http.StatusOK,
			wantBody:   "Balances recalculated successfully",
		},
	}

<<<<<<< HEAD
func TestHandler_recalculateBalances(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		ctx context.Context
		tx  *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
=======
>>>>>>> 0617a2afde94cf5b86ce3dd3494faae90d7b64cd
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/refresh-balance", strings.NewReader(tt.body))

			svc := tt.mock
			if svc == nil {
				svc = &mocks.MockService{}
			}
			h := &Handler{Service: svc}
			h.RecalculateBalances(c)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && !strings.Contains(w.Body.String(), tt.wantBody) {
				t.Errorf("expected body containing %q, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}
