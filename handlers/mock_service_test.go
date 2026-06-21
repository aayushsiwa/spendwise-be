package handlers

import (
	"context"
	"database/sql"
	"io"

	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Compile-time check that mockService implements services.Service.
var _ services.Service = (*mockService)(nil)

// mockService is a configurable stub for services.Service used in handler tests.
// Each field is a function that can be replaced per-test to inject specific behaviour.
type mockService struct {
	createRecordFn      func(ctx context.Context, rec *models.Record) error
	getRecordFn         func(ctx context.Context, id string) (*models.Record, error)
	getRecordsFn        func(ctx context.Context, whereClause string, filterArgs []any, limit, offset int) ([]models.Record, int, error)
	getGroupedRecordsFn func(ctx context.Context, groupBy, whereClause string, filterArgs []any) ([]models.GroupedRecord, error)
	patchRecordFn       func(ctx context.Context, id string, req *models.UpdateRecordRequest) error
	deleteRecordFn      func(ctx context.Context, id string) (int64, error)
	exportRecordsFn     func(ctx context.Context) (*sql.Rows, error)
	importCSVFn         func(ctx context.Context, src io.Reader) (int, int, error)
	importJSONFn        func(ctx context.Context, records []models.Record) (int, error)
	refreshBalancesFn   func(ctx context.Context) error
	updateSummaryFn     func(ctx context.Context) error
	getSummaryFn        func(ctx context.Context, from, to, categoryFilter, typeFilter string) (*models.Summary, error)
	createCategoriesFn  func(ctx context.Context, categories []models.Category) ([]models.Category, error)
	getCategoriesFn     func(ctx context.Context) ([]models.Category, error)
	updateCategoryFn    func(ctx context.Context, id string, cat *models.Category) error
	deleteCategoryFn    func(ctx context.Context, id string) error
}

func (m *mockService) CreateRecord(ctx context.Context, rec *models.Record) error {
	if m.createRecordFn != nil {
		return m.createRecordFn(ctx, rec)
	}
	return nil
}

func (m *mockService) GetRecord(ctx context.Context, id string) (*models.Record, error) {
	if m.getRecordFn != nil {
		return m.getRecordFn(ctx, id)
	}
	return &models.Record{ID: id}, nil
}

func (m *mockService) GetRecords(ctx context.Context, whereClause string, filterArgs []any, limit, offset int) ([]models.Record, int, error) {
	if m.getRecordsFn != nil {
		return m.getRecordsFn(ctx, whereClause, filterArgs, limit, offset)
	}
	return []models.Record{}, 0, nil
}

func (m *mockService) GetGroupedRecords(ctx context.Context, groupBy, whereClause string, filterArgs []any) ([]models.GroupedRecord, error) {
	if m.getGroupedRecordsFn != nil {
		return m.getGroupedRecordsFn(ctx, groupBy, whereClause, filterArgs)
	}
	return []models.GroupedRecord{}, nil
}

func (m *mockService) PatchRecord(ctx context.Context, id string, req *models.UpdateRecordRequest) error {
	if m.patchRecordFn != nil {
		return m.patchRecordFn(ctx, id, req)
	}
	return nil
}

func (m *mockService) DeleteRecord(ctx context.Context, id string) (int64, error) {
	if m.deleteRecordFn != nil {
		return m.deleteRecordFn(ctx, id)
	}
	return 1, nil
}

func (m *mockService) ExportRecords(ctx context.Context) (*sql.Rows, error) {
	if m.exportRecordsFn != nil {
		return m.exportRecordsFn(ctx)
	}
	return nil, nil
}

func (m *mockService) ImportCSV(ctx context.Context, src io.Reader) (int, int, error) {
	if m.importCSVFn != nil {
		return m.importCSVFn(ctx, src)
	}
	return 0, 0, nil
}

func (m *mockService) ImportJSON(ctx context.Context, records []models.Record) (int, error) {
	if m.importJSONFn != nil {
		return m.importJSONFn(ctx, records)
	}
	return len(records), nil
}

func (m *mockService) RefreshBalances(ctx context.Context) error {
	if m.refreshBalancesFn != nil {
		return m.refreshBalancesFn(ctx)
	}
	return nil
}

func (m *mockService) UpdateSummary(ctx context.Context) error {
	if m.updateSummaryFn != nil {
		return m.updateSummaryFn(ctx)
	}
	return nil
}

func (m *mockService) GetSummary(ctx context.Context, from, to, categoryFilter, typeFilter string) (*models.Summary, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, from, to, categoryFilter, typeFilter)
	}
	return &models.Summary{}, nil
}

func (m *mockService) CreateCategories(ctx context.Context, categories []models.Category) ([]models.Category, error) {
	if m.createCategoriesFn != nil {
		return m.createCategoriesFn(ctx, categories)
	}
	return categories, nil
}

func (m *mockService) GetCategories(ctx context.Context) ([]models.Category, error) {
	if m.getCategoriesFn != nil {
		return m.getCategoriesFn(ctx)
	}
	return []models.Category{}, nil
}

func (m *mockService) UpdateCategory(ctx context.Context, id string, cat *models.Category) error {
	if m.updateCategoryFn != nil {
		return m.updateCategoryFn(ctx, id, cat)
	}
	return nil
}

func (m *mockService) DeleteCategory(ctx context.Context, id string) error {
	if m.deleteCategoryFn != nil {
		return m.deleteCategoryFn(ctx, id)
	}
	return nil
}