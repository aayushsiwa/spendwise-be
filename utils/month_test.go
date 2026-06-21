package utils

import "testing"

func TestNextMonth(t *testing.T) {
	type args struct {
		m string
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
			if got := NextMonth(tt.args.m); got != tt.want {
				t.Errorf("NextMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}
