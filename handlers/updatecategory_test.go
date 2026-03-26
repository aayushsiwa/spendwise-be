package handlers

import "testing"

func TestHandler_UpdateCategory(t *testing.T) {
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
			h.UpdateCategory(tt.args.c)
		})
	}
}
