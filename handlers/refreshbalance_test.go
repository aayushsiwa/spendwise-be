package handlers

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestHandler_RecalculateBalances(t *testing.T) {
	type fields struct {
		DB *gorm.DB
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
			h.RecalculateBalances(tt.args.c)
		})
	}
}

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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				DB: tt.fields.DB,
			}
			if err := h.recalculateBalances(tt.args.ctx, tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("recalculateBalances() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
