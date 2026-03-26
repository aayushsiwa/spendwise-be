package handlers

import (
	"aayushsiwa/expense-tracker/models"
	"database/sql"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_GetSummary(t *testing.T) {
	type fields struct {
		DB *sql.DB
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
				DB: tt.fields.DB,
			}
			h.GetSummary(tt.args.c)
		})
	}
}

func TestHandler_UpdateSummary(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				DB: tt.fields.DB,
			}
			if err := h.UpdateSummary(); (err != nil) != tt.wantErr {
				t.Errorf("UpdateSummary() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_buildSummaryWhereClause(t *testing.T) {
	type args struct {
		q *models.QueryParams
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := buildSummaryWhereClause(tt.args.q)
			if got != tt.want {
				t.Errorf("buildSummaryWhereClause() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("buildSummaryWhereClause() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_getSummaryInDateRange(t *testing.T) {
	type args struct {
		h    *Handler
		c    *gin.Context
		from string
		to   string
	}
	tests := []struct {
		name string
		args args
		want *models.Summary
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSummaryInDateRange(tt.args.h, tt.args.c, tt.args.from, tt.args.to); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSummaryInDateRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_monthRangeFromQuery(t *testing.T) {
	type args struct {
		q *models.QueryParams
	}
	tests := []struct {
		name     string
		args     args
		wantFrom string
		wantTo   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrom, gotTo := monthRangeFromQuery(tt.args.q)
			if gotFrom != tt.wantFrom {
				t.Errorf("monthRangeFromQuery() gotFrom = %v, want %v", gotFrom, tt.wantFrom)
			}
			if gotTo != tt.wantTo {
				t.Errorf("monthRangeFromQuery() gotTo = %v, want %v", gotTo, tt.wantTo)
			}
		})
	}
}
