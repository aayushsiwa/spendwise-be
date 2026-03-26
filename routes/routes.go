package routes

import (
	"aayushsiwa/expense-tracker/db"
	"aayushsiwa/expense-tracker/handlers"

	"github.com/gin-gonic/gin"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc gin.HandlerFunc
}

type Routes []Route

func NewRoutes(h *handlers.Handler) Routes {
	return Routes{
		// Records
		{
			Name:        "GetRecords",
			Method:      "GET",
			Pattern:     "/records",
			HandlerFunc: h.GetRecords,
		},
		{
			Name:        "CreateRecord",
			Method:      "POST",
			Pattern:     "/records",
			HandlerFunc: h.CreateRecord,
		},
		{
			Name:        "GetRecord",
			Method:      "GET",
			Pattern:     "/records/:id",
			HandlerFunc: h.GetRecord,
		},
		{
			Name:        "PatchRecord",
			Method:      "PATCH",
			Pattern:     "/records/:id",
			HandlerFunc: h.PatchRecord,
		}, {
			Name:        "DeleteRecord",
			Method:      "DELETE",
			Pattern:     "/records/:id",
			HandlerFunc: h.DeleteRecord,
		},
		// Export
		{
			Name:        "ExportCSV",
			Method:      "GET",
			Pattern:     "/export/csv",
			HandlerFunc: h.ExportCSV,
		},
		// Summary
		{
			Name:        "GetSummary",
			Method:      "GET",
			Pattern:     "/summary",
			HandlerFunc: h.GetSummary,
		},
		{
			Name:        "GetSummaryForFilters",
			Method:      "GET",
			Pattern:     "/summary/filter",
			HandlerFunc: h.GetSummaryForFilters,
		},
		{
			Name:        "GetSummaryForFilter",
			Method:      "GET",
			Pattern:     "/summary/:filter/:value",
			HandlerFunc: h.GetSummaryForFilter,
		},
		// Import
		{
			Name:        "ImportCSV",
			Method:      "POST",
			Pattern:     "/import/csv",
			HandlerFunc: h.ImportCSV,
		},
		{
			Name:        "ImportJSON",
			Method:      "POST",
			Pattern:     "/import/json",
			HandlerFunc: h.ImportJSON,
		},
		// Categories
		{
			Name:        "GetCategories",
			Method:      "GET",
			Pattern:     "/categories",
			HandlerFunc: h.GetCategories,
		},
		{
			Name:        "CreateCategory",
			Method:      "POST",
			Pattern:     "/categories",
			HandlerFunc: h.CreateCategories,
		},
		{
			Name:        "UpdateCategory",
			Method:      "PATCH",
			Pattern:     "/categories/:id",
			HandlerFunc: h.UpdateCategory,
		},
		{
			Name:        "DeleteCategory",
			Method:      "DELETE",
			Pattern:     "/categories/:id",
			HandlerFunc: h.DeleteCategory,
		},
		// Health Check
		{
			Name:    "HealthCheck",
			Method:  "GET",
			Pattern: "/health",
			HandlerFunc: func(c *gin.Context) {
				if err := db.HealthCheck(); err != nil {
					c.JSON(503, gin.H{"status": "unhealthy", "error": err.Error()})
					return
				}
				c.JSON(200, gin.H{"status": "healthy"})
			},
		},
		{
			Name:        "RefreshBalance",
			Method:      "POST",
			Pattern:     "/refresh",
			HandlerFunc: h.RecalculateBalances,
		},
	}
}

func AttachRoutes(server *gin.RouterGroup, routes Routes) {
	for _, route := range routes {
		server.Handle(route.Method, route.Pattern, route.HandlerFunc)
	}
}
