package routes

import (
	"net/http"

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

func NewRoutes() Routes {
	return Routes{
		// Records
		{
			Name:        "GetRecords",
			Method:      "GET",
			Pattern:     "/records",
			HandlerFunc: handlers.GetRecords,
		},
		{
			Name:        "CreateRecord",
			Method:      "POST",
			Pattern:     "/records",
			HandlerFunc: handlers.CreateRecord,
		},
		{
			Name:        "GetRecord",
			Method:      "GET",
			Pattern:     "/records/:id",
			HandlerFunc: handlers.GetRecord,
		},
		{
			Name:        "PatchRecord",
			Method:      "PATCH",
			Pattern:     "/records/:id",
			HandlerFunc: handlers.PatchRecord,
		}, {
			Name:        "DeleteRecord",
			Method:      "DELETE",
			Pattern:     "/records/:id",
			HandlerFunc: handlers.DeleteRecord,
		},
		// Export
		{
			Name:        "ExportCSV",
			Method:      "GET",
			Pattern:     "/export/csv",
			HandlerFunc: handlers.ExportCSV,
		},
		// Summary
		{
			Name:        "GetSummary",
			Method:      "GET",
			Pattern:     "/summary",
			HandlerFunc: handlers.GetSummary,
		},
		{
			Name:        "GetSummaryForFilters",
			Method:      "GET",
			Pattern:     "/summary/filter",
			HandlerFunc: handlers.GetSummaryForFilters,
		},
		{
			Name:        "GetSummaryForFilters",
			Method:      "GET",
			Pattern:     "/summary/:filter/:value",
			HandlerFunc: handlers.GetSummaryForFilter,
		},
		// Import
		{
			Name:        "ImportCSV",
			Method:      "POST",
			Pattern:     "/import/csv",
			HandlerFunc: handlers.ImportCSV,
		},
		{
			Name:        "ImportJSON",
			Method:      "POST",
			Pattern:     "/import/json",
			HandlerFunc: handlers.ImportJSON,
		},
		// Categories
		{
			Name:        "GetCategories",
			Method:      "GET",
			Pattern:     "/categories",
			HandlerFunc: handlers.GetCategories,
		},
		{
			Name:        "CreateCategory",
			Method:      "POST",
			Pattern:     "/categories",
			HandlerFunc: handlers.CreateCategories,
		},
		{
			Name:        "UpdateCategory",
			Method:      "PATCH",
			Pattern:     "/categories/:id",
			HandlerFunc: handlers.UpdateCategory,
		},
		{
			Name:        "DeleteCategory",
			Method:      "DELETE",
			Pattern:     "/categories/:id",
			HandlerFunc: handlers.DeleteCategory,
		},
		// Category Records
		{
			Name:        "GetCategoryRecords",
			Method:      "GET",
			Pattern:     "/categories/:id",
			HandlerFunc: handlers.GetCategoryRecords,
		},
		// Health Check
		{
			Name:    "HealthCheck",
			Method:  "GET",
			Pattern: "/health",
			HandlerFunc: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			},
		},
	}
}

func AttachRoutes(server *gin.RouterGroup, routes Routes) {
	for _, route := range routes {
		server.Handle(route.Method, route.Pattern, route.HandlerFunc)
	}
}
