package utils

import "testing"

func TestParseDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		want    string
		wantErr bool
	}{
		{"ISO format", "2023-07-25", "2023-07-25", false},
		{"DD-MM-YYYY", "25-07-2023", "2023-07-25", false},
		{"DD/MM/YYYY", "25/07/2023", "2023-07-25", false},
		{"MM/DD/YYYY", "07/25/2023", "2023-07-25", false},
		{"M/D/YYYY", "7/25/2023", "2023-07-25", false},
		{"2 Jan 2006", "25 Jul 2023", "2023-07-25", false},
		{"2 Jan, 2006", "25 Jul, 2023", "2023-07-25", false},
		{"2 January 2006", "25 July 2023", "2023-07-25", false},
		{"January 2, 2006", "July 25, 2023", "2023-07-25", false},
		{"Jan 2, 2006", "Jul 25, 2023", "2023-07-25", false},
		{"YYYY/MM/DD", "2023/07/25", "2023-07-25", false},
		{"YYYY.MM.DD", "2023.07.25", "2023-07-25", false},
		{"DD.MM.YYYY", "25.07.2023", "2023-07-25", false},
		{"DD-MM-YYYY with time", "16-06-2026 12:14:34", "2026-06-16", false},
		{"DD/MM/YYYY with time", "16/06/2026 12:14:34", "2026-06-16", false},
		{"whitespace trimmed", "  2023-07-25  ", "2023-07-25", false},
		{"invalid format", "not-a-date", "", true},
		{"empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDate(tt.dateStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseDate() = %q, want %q", got, tt.want)
			}
		})
	}
}
