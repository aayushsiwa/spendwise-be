package handlers

import "testing"

func Test_nextMonth(t *testing.T) {
	tests := []struct {
		name string
		m    string
		want string
	}{
		{name: "jan to feb", m: "2024-01", want: "2024-02"},
		{name: "dec to jan next year", m: "2024-12", want: "2025-01"},
		{name: "feb non-leap to mar", m: "2023-02", want: "2023-03"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextMonth(tt.m); got != tt.want {
				t.Errorf("nextMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}
