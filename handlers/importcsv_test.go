package handlers

import (
	"database/sql"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandler_ImportCSV(t *testing.T) {
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
			h.ImportCSV(tt.args.c)
		})
	}
}

func Test_abs(t *testing.T) {
	type args struct {
		f float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := abs(tt.args.f); got != tt.want {
				t.Errorf("abs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_insertBatch(t *testing.T) {
	type args struct {
		tx    *sql.Tx
		batch []recordRow
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := insertBatch(tt.args.tx, tt.args.batch); (err != nil) != tt.wantErr {
				t.Errorf("insertBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_normalize(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalize(tt.args.s); got != tt.want {
				t.Errorf("normalize() = %v, want %v", got, tt.want)
			}
		})
	}
}
