package handlers

import (
	"aayushsiwa/expense-tracker/models"
	"database/sql"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_GetRecords(t *testing.T) {
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
		want1 []interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := buildWhereClause(tt.args.q)
			if got != tt.want {
				t.Errorf("buildWhereClause() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("buildWhereClause() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
