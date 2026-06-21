package routes

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"aayushsiwa/expense-tracker/db"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

func TestNewRoutes(t *testing.T) {
	routes := NewRoutes(nil)

	if len(routes) != 15 {
		t.Fatalf("expected 15 routes, got %d", len(routes))
	}

	expected := []struct {
		Name    string
		Method  string
		Pattern string
	}{
		{Name: "GetRecords", Method: "GET", Pattern: "/records"},
		{Name: "CreateRecord", Method: "POST", Pattern: "/records"},
		{Name: "GetRecord", Method: "GET", Pattern: "/records/:id"},
		{Name: "PatchRecord", Method: "PATCH", Pattern: "/records/:id"},
		{Name: "DeleteRecord", Method: "DELETE", Pattern: "/records/:id"},
		{Name: "ExportCSV", Method: "GET", Pattern: "/export/csv"},
		{Name: "GetSummary", Method: "GET", Pattern: "/summary"},
		{Name: "ImportCSV", Method: "POST", Pattern: "/import/csv"},
		{Name: "ImportJSON", Method: "POST", Pattern: "/import/json"},
		{Name: "GetCategories", Method: "GET", Pattern: "/categories"},
		{Name: "CreateCategory", Method: "POST", Pattern: "/categories"},
		{Name: "UpdateCategory", Method: "PATCH", Pattern: "/categories/:id"},
		{Name: "DeleteCategory", Method: "DELETE", Pattern: "/categories/:id"},
		{Name: "HealthCheck", Method: "GET", Pattern: "/health"},
		{Name: "RefreshBalance", Method: "POST", Pattern: "/refresh"},
	}

	for i, exp := range expected {
		t.Run(exp.Name, func(t *testing.T) {
			if routes[i].Name != exp.Name {
				t.Errorf("route %d Name = %q, want %q", i, routes[i].Name, exp.Name)
			}
			if routes[i].Method != exp.Method {
				t.Errorf("route %d Method = %q, want %q", i, routes[i].Method, exp.Method)
			}
			if routes[i].Pattern != exp.Pattern {
				t.Errorf("route %d Pattern = %q, want %q", i, routes[i].Pattern, exp.Pattern)
			}
			if routes[i].HandlerFunc == nil {
				t.Errorf("route %d HandlerFunc is nil", i)
			}
		})
	}
}

func TestAttachRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	group := engine.Group("/api/v1")

	routes := Routes{
		{
			Name:    "Ping",
			Method:  "GET",
			Pattern: "/ping",
			HandlerFunc: func(c *gin.Context) {
				c.String(http.StatusOK, "pong")
			},
		},
		{
			Name:    "Echo",
			Method:  "POST",
			Pattern: "/echo",
			HandlerFunc: func(c *gin.Context) {
				c.String(http.StatusOK, "echo")
			},
		},
	}

	AttachRoutes(group, routes)

	t.Run("GET /api/v1/ping", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "pong" {
			t.Errorf("expected body %q, got %q", "pong", w.Body.String())
		}
	})

	t.Run("POST /api/v1/echo", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/echo", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		if w.Body.String() != "echo" {
			t.Errorf("expected body %q, got %q", "echo", w.Body.String())
		}
	})

	t.Run("405 wrong method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ping", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404 for wrong method, got %d", w.Code)
		}
	})

	t.Run("404 unknown route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404 for unknown route, got %d", w.Code)
		}
	})
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original DB
	origDB := db.DB
	defer func() { db.DB = origDB }()

	t.Run("unhealthy", func(t *testing.T) {
		db.DB = nil

		r := gin.New()
		routes := NewRoutes(nil)
		var hcRoute Route
		for _, route := range routes {
			if route.Name == "HealthCheck" {
				hcRoute = route
				break
			}
		}
		if hcRoute.HandlerFunc == nil {
			t.Fatal("HealthCheck route not found")
		}

		r.GET(hcRoute.Pattern, hcRoute.HandlerFunc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})

	t.Run("healthy", func(t *testing.T) {
		testDb, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatalf("failed to open test db: %v", err)
		}
		defer func() {
			if err := testDb.Close(); err != nil {
				t.Fatalf("failed to close test db: %v", err)
			}
		}()
		db.DB = testDb

		r := gin.New()
		routes := NewRoutes(nil)
		var hcRoute Route
		for _, route := range routes {
			if route.Name == "HealthCheck" {
				hcRoute = route
				break
			}
		}

		r.GET(hcRoute.Pattern, hcRoute.HandlerFunc)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}
