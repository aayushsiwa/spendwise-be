package handlers

import "testing"

func TestHandler_GenerateCustomID(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		date string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				DB: tt.fields.DB,
			}
			got, err := h.GenerateCustomID(tt.args.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateCustomID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GenerateCustomID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
