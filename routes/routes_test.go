package routes

import (
	"aayushsiwa/expense-tracker/handlers"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAttachRoutes(t *testing.T) {
	type args struct {
		server *gin.RouterGroup
		routes Routes
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AttachRoutes(tt.args.server, tt.args.routes)
		})
	}
}

func TestNewRoutes(t *testing.T) {
	type args struct {
		h *handlers.Handler
	}
	tests := []struct {
		name string
		args args
		want Routes
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRoutes(tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRoutes() = %v, want %v", got, tt.want)
			}
		})
	}
}
