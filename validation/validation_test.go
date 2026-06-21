package validation

import (
	"testing"

	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator() returned nil")
	}
	if v.errors == nil {
		t.Fatal("errors slice not initialized")
	}
}

func TestHasErrors_GetErrors(t *testing.T) {
	v := NewValidator()
	if v.HasErrors() {
		t.Error("fresh validator should not have errors")
	}
	if len(v.GetErrors()) != 0 {
		t.Error("fresh validator GetErrors should be empty")
	}

	v.errors = append(v.errors, errors.NewValidationError("x", "err", ""))
	if !v.HasErrors() {
		t.Error("should have errors after adding one")
	}
	if len(v.GetErrors()) != 1 {
		t.Error("GetErrors should return 1 error")
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name     string
		idStr    string
		wantID   string
		wantErrs int
	}{
		{"empty returns error", "", "", 1},
		{"valid ID returns as-is", "abc123", "abc123", 0},
		{"whitespace treated as valid (only checks empty)", "   ", "   ", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			gotID, errs := v.ValidateID(tt.idStr)
			if gotID != tt.wantID {
				t.Errorf("got id %q, want %q", gotID, tt.wantID)
			}
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %+v", len(errs), tt.wantErrs, errs)
			}
		})
	}
}

func TestValidateRecord(t *testing.T) {
	tests := []struct {
		name     string
		record   *models.Record
		wantErrs int
		check    func(t *testing.T, errs errors.ValidationErrors)
	}{
		{
			name:     "missing required fields",
			record:   &models.Record{},
			wantErrs: 3,
			check: func(t *testing.T, errs errors.ValidationErrors) {
				fields := map[string]bool{}
				for _, e := range errs {
					fields[e.Field] = true
				}
				if !fields["date"] {
					t.Error("missing date error")
				}
				if !fields["category"] {
					t.Error("missing category error")
				}
				if !fields["amount"] {
					t.Error("missing amount error")
				}
			},
		},
		{
			name: "invalid date format",
			record: &models.Record{
				Date: "01-15-2024", Category: "Food", Amount: 10, Description: "x",
			},
			wantErrs: 1,
		},
		{
			name: "zero amount",
			record: &models.Record{
				Date: "2024-01-15", Category: "Food", Amount: 0, Description: "x",
			},
			wantErrs: 1,
		},
		{
			name: "negative amount",
			record: &models.Record{
				Date: "2024-01-15", Category: "Food", Amount: -5, Description: "x",
			},
			wantErrs: 1,
		},
		{
			name: "description too long",
			record: &models.Record{
				Date: "2024-01-15", Category: "Food", Amount: 10,
				Description: string(make([]byte, 256)),
			},
			wantErrs: 1,
		},
		{
			name: "note too long",
			record: &models.Record{
				Date: "2024-01-15", Category: "Food", Amount: 10, Description: "x",
				Note: string(make([]byte, 1001)),
			},
			wantErrs: 1,
		},
		{
			name: "valid record",
			record: &models.Record{
				Date: "2024-01-15", Category: "Food", Amount: 10, Description: "groceries",
			},
			wantErrs: 0,
		},
		{
			name: "valid record with note",
			record: &models.Record{
				Date: "2024-01-15", Category: "Food", Amount: 10, Description: "groceries",
				Note: "weekly shop",
			},
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			errs := v.ValidateRecord(tt.record)
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %+v", len(errs), tt.wantErrs, errs)
			}
			if tt.check != nil {
				tt.check(t, errs)
			}
		})
	}
}

func TestValidatePatchRecord(t *testing.T) {
	tests := []struct {
		name     string
		req      *models.UpdateRecordRequest
		wantErrs int
	}{
		{
			name:     "empty patch is valid",
			req:      &models.UpdateRecordRequest{},
			wantErrs: 0,
		},
		{
			name: "invalid date",
			req: &models.UpdateRecordRequest{
				Date: new("01-15-2024"),
			},
			wantErrs: 1,
		},
		{
			name: "valid date",
			req: &models.UpdateRecordRequest{
				Date: new("2024-01-15"),
			},
			wantErrs: 0,
		},
		{
			name: "negative amount",
			req: &models.UpdateRecordRequest{
				Amount: new(float64(-1)),
			},
			wantErrs: 1,
		},
		{
			name: "zero amount",
			req: &models.UpdateRecordRequest{
				Amount: new(float64(0)),
			},
			wantErrs: 1,
		},
		{
			name: "valid amount",
			req: &models.UpdateRecordRequest{
				Amount: new(float64(50)),
			},
			wantErrs: 0,
		},
		{
			name: "description too long",
			req: &models.UpdateRecordRequest{
				Description: new(string(make([]byte, 256))),
			},
			wantErrs: 1,
		},
		{
			name: "note too long",
			req: &models.UpdateRecordRequest{
				Note: new(string(make([]byte, 1001))),
			},
			wantErrs: 1,
		},
		{
			name: "valid patch all fields",
			req: &models.UpdateRecordRequest{
				Date:        new("2024-06-01"),
				Description: new("new desc"),
				Amount:      new(99.99),
				Note:        new("updated note"),
			},
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			errs := v.ValidatePatchRecord(tt.req)
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %+v", len(errs), tt.wantErrs, errs)
			}
		})
	}
}

func TestValidateCategory(t *testing.T) {
	tests := []struct {
		name     string
		cat      *models.Category
		wantErrs int
	}{
		{
			name:     "empty name",
			cat:      &models.Category{},
			wantErrs: 1,
		},
		{
			name:     "invalid characters in name",
			cat:      &models.Category{Name: "Food!!!"},
			wantErrs: 1,
		},
		{
			name:     "name too long",
			cat:      &models.Category{Name: string(make([]byte, 51))},
			wantErrs: 2,
		},
		{
			name:     "invalid hex color",
			cat:      &models.Category{Name: "Food", Color: "red"},
			wantErrs: 1,
		},
		{
			name:     "icon too long",
			cat:      &models.Category{Name: "Food", Icon: string(make([]byte, 51))},
			wantErrs: 1,
		},
		{
			name:     "valid category",
			cat:      &models.Category{Name: "Food", Icon: "pizza", Color: "#FF0000"},
			wantErrs: 0,
		},
		{
			name:     "valid category no icon or color",
			cat:      &models.Category{Name: "Transport"},
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			errs := v.ValidateCategory(tt.cat)
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %+v", len(errs), tt.wantErrs, errs)
			}
		})
	}
}

func TestEnum(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		allowed []string
		wantErr bool
	}{
		{"value in list", "income", []string{"income", "expense", "transfer"}, false},
		{"value not in list", "invalid", []string{"income", "expense", "transfer"}, true},
		{"empty allowed list", "x", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.enum("type", tt.value, tt.allowed, "invalid type")
			hasErr := v.HasErrors()
			if hasErr != tt.wantErr {
				t.Errorf("HasErrors() = %v, want %v; errors=%+v", hasErr, tt.wantErr, v.errors)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		min     int
		wantErr bool
	}{
		{"shorter than min", "ab", 3, true},
		{"exactly min", "abc", 3, false},
		{"longer than min", "abcd", 3, false},
		{"empty with min=1", "", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.minLength("f", tt.value, tt.min, "too short")
			if v.HasErrors() != tt.wantErr {
				t.Errorf("minLength(%q, %d) → HasErrors()=%v, want %v", tt.value, tt.min, v.HasErrors(), tt.wantErr)
			}
		})
	}
}
