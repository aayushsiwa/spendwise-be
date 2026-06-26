package handlers

import (
	"strings"
	"testing"

	"aayushsiwa/expense-tracker/mocks"
)

func TestGenerateCustomID(t *testing.T) {
	tests := []struct {
		name  string
		date  string
		check func(t *testing.T, got string, err error)
	}{
		{
			name: "returns non-empty string",
			date: "2024-01-15",
			check: func(t *testing.T, got string, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if got == "" {
					t.Error("expected non-empty ID")
				}
			},
		},
		{
			name: "no error",
			date: "",
			check: func(t *testing.T, got string, err error) {
				if err != nil {
					t.Errorf("GenerateCustomID should never return an error, got: %v", err)
				}
			},
		},
		{
			name: "returns unique IDs",
			date: "2024-01-01",
			check: func(t *testing.T, got string, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				h := &Handler{Service: &mocks.MockService{}}
				id2, err2 := h.GenerateCustomID("2024-01-01")
				if err2 != nil {
					t.Errorf("expected no error on second call, got %v", err2)
				}
				if got == id2 {
					t.Error("GenerateCustomID should return unique IDs on successive calls")
				}
			},
		},
		{
			name: "date parameter ignored",
			date: "",
			check: func(t *testing.T, got string, err error) {
				h := &Handler{Service: &mocks.MockService{}}
				for _, date := range []string{"", "2024-06-01", "not-a-date", "9999-12-31"} {
					got, err := h.GenerateCustomID(date)
					if err != nil {
						t.Errorf("GenerateCustomID(%q) returned unexpected error: %v", date, err)
					}
					if got == "" {
						t.Errorf("GenerateCustomID(%q) returned empty string", date)
					}
				}
			},
		},
		{
			name: "ID contains only alphanumeric",
			date: "2024-01-15",
			check: func(t *testing.T, got string, err error) {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ123456789"
				for _, ch := range got {
					if !strings.ContainsRune(alphabet, ch) {
						t.Errorf("unexpected character %q in generated ID %q", ch, got)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{Service: &mocks.MockService{}}
			got, err := h.GenerateCustomID(tt.date)
			tt.check(t, got, err)
		})
	}
}

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T)
	}{
		{
			name: "returns handler with given service and initialized internals",
			check: func(t *testing.T) {
				mockSvc := &mocks.MockService{}
				h := NewHandler(mockSvc)
				if h == nil {
					t.Fatal("expected NewHandler to return a non-nil Handler")
				}
				if h.Service != mockSvc {
					t.Errorf("expected Handler to have the provided service, got %v", h.Service)
				}

				if genIDFunc == nil || openFileFunc == nil {
					t.Error("expected internal helper functions to be initialized")
				}

				_, _ = genIDFunc("2024-01-01")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t)
		})
	}
}
