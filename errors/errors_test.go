package errors

import (
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAppError_Error(t *testing.T) {
	type fields struct {
		Type       string
		Message    string
		Details    map[string]interface{}
		StatusCode int
		Err        error
		Context    map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AppError{
				Type:       tt.fields.Type,
				Message:    tt.fields.Message,
				Details:    tt.fields.Details,
				StatusCode: tt.fields.StatusCode,
				Err:        tt.fields.Err,
				Context:    tt.fields.Context,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Log(t *testing.T) {
	type fields struct {
		Type       string
		Message    string
		Details    map[string]interface{}
		StatusCode int
		Err        error
		Context    map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AppError{
				Type:       tt.fields.Type,
				Message:    tt.fields.Message,
				Details:    tt.fields.Details,
				StatusCode: tt.fields.StatusCode,
				Err:        tt.fields.Err,
				Context:    tt.fields.Context,
			}
			e.Log()
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	type fields struct {
		Type       string
		Message    string
		Details    map[string]interface{}
		StatusCode int
		Err        error
		Context    map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AppError{
				Type:       tt.fields.Type,
				Message:    tt.fields.Message,
				Details:    tt.fields.Details,
				StatusCode: tt.fields.StatusCode,
				Err:        tt.fields.Err,
				Context:    tt.fields.Context,
			}
			if err := e.Unwrap(); (err != nil) != tt.wantErr {
				t.Errorf("Unwrap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAppError_WithContext(t *testing.T) {
	type fields struct {
		Type       string
		Message    string
		Details    map[string]interface{}
		StatusCode int
		Err        error
		Context    map[string]interface{}
	}
	type args struct {
		key   string
		value interface{}
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
			e := &AppError{
				Type:       tt.fields.Type,
				Message:    tt.fields.Message,
				Details:    tt.fields.Details,
				StatusCode: tt.fields.StatusCode,
				Err:        tt.fields.Err,
				Context:    tt.fields.Context,
			}
			e.WithContext(tt.args.key, tt.args.value)
		})
	}
}

func TestAppError_WithDetails(t *testing.T) {
	type fields struct {
		Type       string
		Message    string
		Details    map[string]interface{}
		StatusCode int
		Err        error
		Context    map[string]interface{}
	}
	type args struct {
		details map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AppError{
				Type:       tt.fields.Type,
				Message:    tt.fields.Message,
				Details:    tt.fields.Details,
				StatusCode: tt.fields.StatusCode,
				Err:        tt.fields.Err,
				Context:    tt.fields.Context,
			}
			if got := e.WithDetails(tt.args.details); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	type args struct {
		c   *gin.Context
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleError(tt.args.c, tt.args.err)
		})
	}
}

func TestHandleValidationErrors(t *testing.T) {
	type args struct {
		c              *gin.Context
		validationErrs ValidationErrors
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleValidationErrors(tt.args.c, tt.args.validationErrs)
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		errType    string
		message    string
		statusCode int
		err        error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.errType, tt.args.message, tt.args.statusCode, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConflict(t *testing.T) {
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConflict(tt.args.message, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDatabase(t *testing.T) {
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDatabase(tt.args.message, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDatabase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEncryption(t *testing.T) {
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewEncryption(tt.args.message, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEncryption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewForbidden(t *testing.T) {
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewForbidden(tt.args.message, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewForbidden() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewInternal(t *testing.T) {
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInternal(tt.args.message, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInternal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewInvalidInput(t *testing.T) {
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInvalidInput(tt.args.message, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInvalidInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewNotFound(t *testing.T) {
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNotFound(tt.args.message, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUnauthorized(t *testing.T) {
	type args struct {
		message string
		err     error
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUnauthorized(tt.args.message, tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUnauthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewValidation(t *testing.T) {
	type args struct {
		message string
		details map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want *AppError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewValidation(tt.args.message, tt.args.details); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	type args struct {
		field   string
		message string
		value   interface{}
	}
	tests := []struct {
		name string
		args args
		want ValidationError
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewValidationError(tt.args.field, tt.args.message, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name string
		v    ValidationErrors
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidationErrors_ToMap(t *testing.T) {
	tests := []struct {
		name string
		v    ValidationErrors
		want map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.ToMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHandlerName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHandlerName(); got != tt.want {
				t.Errorf("getHandlerName() = %v, want %v", got, tt.want)
			}
		})
	}
}
