package validation

import (
	"aayushsiwa/expense-tracker/errors"
	"aayushsiwa/expense-tracker/models"
	"reflect"
	"testing"
)

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name string
		want *Validator
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewValidator(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidator_GetErrors(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	tests := []struct {
		name   string
		fields fields
		want   errors.ValidationErrors
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				errors: tt.fields.errors,
			}
			if got := v.GetErrors(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidator_HasErrors(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				errors: tt.fields.errors,
			}
			if got := v.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidator_ValidateCategory(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		category *models.Category
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   errors.ValidationErrors
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				errors: tt.fields.errors,
			}
			if got := v.ValidateCategory(tt.args.category); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidator_ValidateID(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		idStr string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
		want1  errors.ValidationErrors
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				errors: tt.fields.errors,
			}
			got, got1 := v.ValidateID(tt.args.idStr)
			if got != tt.want {
				t.Errorf("ValidateID() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ValidateID() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestValidator_ValidateRecord(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		record *models.Record
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   errors.ValidationErrors
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Validator{
				errors: tt.fields.errors,
			}
			if got := v.ValidateRecord(tt.args.record); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidator_dateFormat(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		field   string
		value   string
		message string
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
			v := &Validator{
				errors: tt.fields.errors,
			}
			v.dateFormat(tt.args.field, tt.args.value, tt.args.message)
		})
	}
}

func TestValidator_enum(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		field   string
		value   string
		allowed []string
		message string
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
			v := &Validator{
				errors: tt.fields.errors,
			}
			v.enum(tt.args.field, tt.args.value, tt.args.allowed, tt.args.message)
		})
	}
}

func TestValidator_maxLength(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		field   string
		value   string
		max     int
		message string
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
			v := &Validator{
				errors: tt.fields.errors,
			}
			v.maxLength(tt.args.field, tt.args.value, tt.args.max, tt.args.message)
		})
	}
}

func TestValidator_minLength(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		field   string
		value   string
		min     int
		message string
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
			v := &Validator{
				errors: tt.fields.errors,
			}
			v.minLength(tt.args.field, tt.args.value, tt.args.min, tt.args.message)
		})
	}
}

func TestValidator_pattern(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		field   string
		value   string
		pattern string
		message string
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
			v := &Validator{
				errors: tt.fields.errors,
			}
			v.pattern(tt.args.field, tt.args.value, tt.args.pattern, tt.args.message)
		})
	}
}

func TestValidator_positiveNumber(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		field   string
		value   float64
		message string
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
			v := &Validator{
				errors: tt.fields.errors,
			}
			v.positiveNumber(tt.args.field, tt.args.value, tt.args.message)
		})
	}
}

func TestValidator_required(t *testing.T) {
	type fields struct {
		errors errors.ValidationErrors
	}
	type args struct {
		field   string
		value   string
		message string
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
			v := &Validator{
				errors: tt.fields.errors,
			}
			v.required(tt.args.field, tt.args.value, tt.args.message)
		})
	}
}
