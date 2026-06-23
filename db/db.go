package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"time"

	"aayushsiwa/expense-tracker/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init initializes the database connection and runs schema migrations.
// It reads DB_TYPE (defaulting to "sqlite") and DB_URL or DATABASE_URL environment
// variables to determine database type and connection URL. For SQLite, defaultPath
// is used when no URL environment variable is set. It returns the initialized GORM
// database instance.
func Init(defaultPath string) (*gorm.DB, error) {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}

	var dialector gorm.Dialector

	switch dbType {
	case "postgres":
		if dbURL == "" {
			return nil, errors.New("DB_URL or DATABASE_URL is required for postgres")
		}
		dialector = postgres.Open(dbURL)
	case "mysql":
		if dbURL == "" {
			return nil, errors.New("DB_URL or DATABASE_URL is required for mysql")
		}
		dialector = mysql.Open(dbURL)
	case "sqlite":
		if dbURL == "" {
			dbURL = defaultPath
		}
		dialector = sqlite.Open(dbURL)
	default:
		return nil, errors.New("unsupported database type: " + dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to open database", "type", dbType, "error", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to get raw database instance", "error", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = sqlDB.PingContext(ctx); err != nil {
		slog.ErrorContext(ctx, "Database connection failed", "error", err)
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "Error closing sql db", "error", closeErr)
		}
		return nil, err
	}

	// SQLite specific settings
	if dbType == "sqlite" {
		if _, err := sqlDB.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
			slog.WarnContext(ctx, "Failed to set PRAGMA foreign_keys", "error", err)
		}
		if _, err := sqlDB.ExecContext(ctx, `PRAGMA journal_mode = WAL;`); err != nil {
			slog.WarnContext(ctx, "Failed to set PRAGMA journal_mode", "error", err)
		}
		if _, err := sqlDB.ExecContext(ctx, `PRAGMA synchronous = NORMAL;`); err != nil {
			slog.WarnContext(ctx, "Failed to set PRAGMA synchronous", "error", err)
		}
	}

	// AutoMigrate tables
	err = db.AutoMigrate(
		&models.Category{},
		&models.Record{},
		&models.SummaryDB{},
		&models.SummaryDetailDB{},
	)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to auto-migrate tables", "error", err)
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "Error closing sql db", "error", closeErr)
		}
		return nil, err
	}

	DB = db

	slog.InfoContext(ctx, "Database connected and migrated successfully", "type", dbType)
	return db, nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		slog.InfoContext(context.Background(), "Closing database connection")
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// HealthCheck verifies that the database connection is working.
func HealthCheck() error {
	if DB == nil {
		return errors.New("database not initialized")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}

// GetStats retrieves statistics for the current database connection. If the database is not initialized or retrieval fails, it returns empty statistics.
func GetStats() sql.DBStats {
	if DB == nil {
		return sql.DBStats{}
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return sql.DBStats{}
	}
	return sqlDB.Stats()
}
