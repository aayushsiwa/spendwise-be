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

// Init initializes the database connection with proper error handling and migrations
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
		slog.Error("Failed to open database", "type", dbType, "error", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get raw database instance", "error", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = sqlDB.PingContext(ctx); err != nil {
		slog.Error("Database connection failed", "error", err)
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.Error("Error closing sql db", "error", closeErr)
		}
		return nil, err
	}

	// SQLite specific settings
	if dbType == "sqlite" {
		if _, err := sqlDB.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
			slog.Warn("Failed to set PRAGMA foreign_keys", "error", err)
		}
		if _, err := sqlDB.ExecContext(ctx, `PRAGMA journal_mode = WAL;`); err != nil {
			slog.Warn("Failed to set PRAGMA journal_mode", "error", err)
		}
		if _, err := sqlDB.ExecContext(ctx, `PRAGMA synchronous = NORMAL;`); err != nil {
			slog.Warn("Failed to set PRAGMA synchronous", "error", err)
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
		slog.Error("Failed to auto-migrate tables", "error", err)
		if closeErr := sqlDB.Close(); closeErr != nil {
			slog.Error("Error closing sql db", "error", closeErr)
		}
		return nil, err
	}

	DB = db

	slog.Info("Database connected and migrated successfully", "type", dbType)
	return db, nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		slog.Info("Closing database connection")
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// HealthCheck performs a database health check
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

// GetStats returns database statistics
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
