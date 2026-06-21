package handlers

import (
	"testing"

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
