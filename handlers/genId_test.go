package handlers

import (
	"strings"
	"testing"

	"aayushsiwa/expense-tracker/services"
)

func TestHandler_GenerateCustomID(t *testing.T) {
	type fields struct {
		Service services.Service
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
				Service: tt.fields.Service,
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

func TestGenerateCustomID_ReturnsNonEmptyString(t *testing.T) {
	h := &Handler{Service: &mockService{}}
	got, err := h.GenerateCustomID("2024-01-15")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if got == "" {
		t.Error("expected non-empty ID")
	}
}

func TestGenerateCustomID_NoError(t *testing.T) {
	h := &Handler{Service: &mockService{}}
	_, err := h.GenerateCustomID("")
	if err != nil {
		t.Errorf("GenerateCustomID should never return an error, got: %v", err)
	}
}

func TestGenerateCustomID_ReturnsUniqueIDs(t *testing.T) {
	h := &Handler{Service: &mockService{}}
	id1, _ := h.GenerateCustomID("2024-01-01")
	id2, _ := h.GenerateCustomID("2024-01-01")
	if id1 == id2 {
		t.Error("GenerateCustomID should return unique IDs on successive calls")
	}
}

func TestGenerateCustomID_DateParameterIgnored(t *testing.T) {
	// The current implementation ignores the date parameter — any value
	// (or even empty) should still produce a valid ID without error.
	h := &Handler{Service: &mockService{}}
	for _, date := range []string{"", "2024-06-01", "not-a-date", "9999-12-31"} {
		got, err := h.GenerateCustomID(date)
		if err != nil {
			t.Errorf("GenerateCustomID(%q) returned unexpected error: %v", date, err)
		}
		if got == "" {
			t.Errorf("GenerateCustomID(%q) returned empty string", date)
		}
	}
}

func TestGenerateCustomID_IDContainsOnlyAlphanumeric(t *testing.T) {
	h := &Handler{Service: &mockService{}}
	got, _ := h.GenerateCustomID("2024-01-15")
	// shortuuid uses a base57 alphabet (alphanumeric without ambiguous characters)
	const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ123456789"
	for _, ch := range got {
		if !strings.ContainsRune(alphabet, ch) {
			t.Errorf("unexpected character %q in generated ID %q", ch, got)
		}
	}
}
