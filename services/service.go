package services

import (
	"context"
	"io"

	"aayushsiwa/expense-tracker/models"

	"gorm.io/gorm"
)

type Service interface {
	CreateRecord(ctx context.Context, rec *models.Record) error
	GetRecord(ctx context.Context, id string) (*models.Record, error)
	GetRecords(ctx context.Context, params *models.QueryParams) ([]models.Record, int, error)
	GetGroupedRecords(ctx context.Context, params *models.QueryParams) ([]models.GroupedRecord, error)
	PatchRecord(ctx context.Context, id string, req *models.UpdateRecordRequest) error
	DeleteRecord(ctx context.Context, id string) (int64, error)

	ExportRecords(ctx context.Context) ([]models.Record, error)

	ImportCSV(ctx context.Context, src io.Reader) (imported, skipped int, err error)
	ImportJSON(ctx context.Context, records []models.Record) (imported int, err error)

	RefreshBalances(ctx context.Context) error

	UpdateSummary(ctx context.Context) error
	GetSummary(ctx context.Context, from, to, categoryFilter, typeFilter string) (*models.Summary, error)

	CreateCategories(ctx context.Context, categories []models.Category) ([]models.Category, error)
	GetCategories(ctx context.Context) ([]models.Category, error)
	UpdateCategory(ctx context.Context, id string, cat *models.Category) error
	DeleteCategory(ctx context.Context, id string) error
}

type RecordService struct {
	db *gorm.DB
}

// NewRecordService creates a new RecordService with the provided database connection.
func NewRecordService(db *gorm.DB) *RecordService {
	return &RecordService{db: db}
}
