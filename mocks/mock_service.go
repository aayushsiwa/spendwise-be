package mocks

import (
	"context"
	"io"

	"aayushsiwa/expense-tracker/models"
	"aayushsiwa/expense-tracker/services"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

var _ services.Service = (*MockService)(nil)

type MockService struct {
	CreateRecordErr      error
	GetRecordErr         error
	GetRecordsErr        error
	GetGroupedRecordsErr error
	PatchRecordErr       error
	DeleteRecordErr      error
	ExportRecordsErr     error
	ImportCSVErr         error
	ImportJSONErr        error
	RefreshBalancesErr   error
	UpdateSummaryErr     error
	GetSummaryErr        error
	CreateCategoriesErr  error
	GetCategoriesErr     error
	UpdateCategoryErr    error
	DeleteCategoryErr    error

	GetRecordResult         *models.Record
	GetRecordsResult        []models.Record
	GetRecordsTotalCount    int
	GetGroupedRecordsResult []models.GroupedRecord
	ExportRecordsResult     []models.Record
	GetSummaryResult        *models.Summary
	CreateCategoriesResult  []models.Category
	GetCategoriesResult     []models.Category

	CreateRecordFn      func(ctx context.Context, rec *models.Record) error
	GetRecordFn         func(ctx context.Context, id string) (*models.Record, error)
	GetRecordsFn        func(ctx context.Context, params *models.QueryParams) ([]models.Record, int, error)
	GetGroupedRecordsFn func(ctx context.Context, params *models.QueryParams) ([]models.GroupedRecord, error)
	PatchRecordFn       func(ctx context.Context, id string, req *models.UpdateRecordRequest) error
	DeleteRecordFn      func(ctx context.Context, id string) (int64, error)
	ExportRecordsFn     func(ctx context.Context) ([]models.Record, error)
	ImportCSVFn         func(ctx context.Context, src io.Reader) (int, int, error)
	ImportJSONFn        func(ctx context.Context, records []models.Record) (int, error)
	RefreshBalancesFn   func(ctx context.Context) error
	UpdateSummaryFn     func(ctx context.Context) error
	GetSummaryFn        func(ctx context.Context, from, to, categoryFilter, typeFilter string) (*models.Summary, error)
	CreateCategoriesFn  func(ctx context.Context, categories []models.Category) ([]models.Category, error)
	GetCategoriesFn     func(ctx context.Context) ([]models.Category, error)
	UpdateCategoryFn    func(ctx context.Context, id string, cat *models.Category) error
	DeleteCategoryFn    func(ctx context.Context, id string) error
}

func (m *MockService) CreateRecord(ctx context.Context, rec *models.Record) error {
	if m.CreateRecordFn != nil {
		return m.CreateRecordFn(ctx, rec)
	}
	return m.CreateRecordErr
}

func (m *MockService) GetRecord(ctx context.Context, id string) (*models.Record, error) {
	if m.GetRecordFn != nil {
		return m.GetRecordFn(ctx, id)
	}
	if m.GetRecordErr != nil {
		return nil, m.GetRecordErr
	}
	if m.GetRecordResult != nil {
		return m.GetRecordResult, nil
	}
	return &models.Record{ID: id}, nil
}

func (m *MockService) GetRecords(ctx context.Context, params *models.QueryParams) ([]models.Record, int, error) {
	if m.GetRecordsFn != nil {
		return m.GetRecordsFn(ctx, params)
	}
	if m.GetRecordsErr != nil {
		return nil, 0, m.GetRecordsErr
	}
	return m.GetRecordsResult, m.GetRecordsTotalCount, nil
}

func (m *MockService) GetGroupedRecords(ctx context.Context, params *models.QueryParams) ([]models.GroupedRecord, error) {
	if m.GetGroupedRecordsFn != nil {
		return m.GetGroupedRecordsFn(ctx, params)
	}
	if m.GetGroupedRecordsErr != nil {
		return nil, m.GetGroupedRecordsErr
	}
	return m.GetGroupedRecordsResult, nil
}

func (m *MockService) PatchRecord(ctx context.Context, id string, req *models.UpdateRecordRequest) error {
	if m.PatchRecordFn != nil {
		return m.PatchRecordFn(ctx, id, req)
	}
	return m.PatchRecordErr
}

func (m *MockService) DeleteRecord(ctx context.Context, id string) (int64, error) {
	if m.DeleteRecordFn != nil {
		return m.DeleteRecordFn(ctx, id)
	}
	if m.DeleteRecordErr != nil {
		return 0, m.DeleteRecordErr
	}
	return 1, nil
}

func (m *MockService) ExportRecords(ctx context.Context) ([]models.Record, error) {
	if m.ExportRecordsFn != nil {
		return m.ExportRecordsFn(ctx)
	}
	if m.ExportRecordsErr != nil {
		return nil, m.ExportRecordsErr
	}
	return m.ExportRecordsResult, nil
}

func (m *MockService) ImportCSV(ctx context.Context, src io.Reader) (int, int, error) {
	if m.ImportCSVFn != nil {
		return m.ImportCSVFn(ctx, src)
	}
	return 0, 0, m.ImportCSVErr
}

func (m *MockService) ImportJSON(ctx context.Context, records []models.Record) (int, error) {
	if m.ImportJSONFn != nil {
		return m.ImportJSONFn(ctx, records)
	}
	if m.ImportJSONErr != nil {
		return 0, m.ImportJSONErr
	}
	return len(records), nil
}

func (m *MockService) RefreshBalances(ctx context.Context) error {
	if m.RefreshBalancesFn != nil {
		return m.RefreshBalancesFn(ctx)
	}
	return m.RefreshBalancesErr
}

func (m *MockService) UpdateSummary(ctx context.Context) error {
	if m.UpdateSummaryFn != nil {
		return m.UpdateSummaryFn(ctx)
	}
	return m.UpdateSummaryErr
}

func (m *MockService) GetSummary(ctx context.Context, from, to, categoryFilter, typeFilter string) (*models.Summary, error) {
	if m.GetSummaryFn != nil {
		return m.GetSummaryFn(ctx, from, to, categoryFilter, typeFilter)
	}
	if m.GetSummaryErr != nil {
		return nil, m.GetSummaryErr
	}
	if m.GetSummaryResult != nil {
		return m.GetSummaryResult, nil
	}
	return &models.Summary{}, nil
}

func (m *MockService) CreateCategories(ctx context.Context, categories []models.Category) ([]models.Category, error) {
	if m.CreateCategoriesFn != nil {
		return m.CreateCategoriesFn(ctx, categories)
	}
	if m.CreateCategoriesErr != nil {
		return nil, m.CreateCategoriesErr
	}
	if m.CreateCategoriesResult != nil {
		return m.CreateCategoriesResult, nil
	}
	return categories, nil
}

func (m *MockService) GetCategories(ctx context.Context) ([]models.Category, error) {
	if m.GetCategoriesFn != nil {
		return m.GetCategoriesFn(ctx)
	}
	if m.GetCategoriesErr != nil {
		return nil, m.GetCategoriesErr
	}
	return m.GetCategoriesResult, nil
}

func (m *MockService) UpdateCategory(ctx context.Context, id string, cat *models.Category) error {
	if m.UpdateCategoryFn != nil {
		return m.UpdateCategoryFn(ctx, id, cat)
	}
	return m.UpdateCategoryErr
}

func (m *MockService) DeleteCategory(ctx context.Context, id string) error {
	if m.DeleteCategoryFn != nil {
		return m.DeleteCategoryFn(ctx, id)
	}
	return m.DeleteCategoryErr
}
