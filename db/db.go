package db

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"

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
		fallthrough
	default:
		if dbURL == "" {
			dbURL = defaultPath
		}
		dialector = sqlite.Open(dbURL)
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

	if err = sqlDB.Ping(); err != nil {
		slog.Error("Database connection failed", "error", err)
		return nil, err
	}

	// SQLite specific settings
	if dbType == "sqlite" {
		_, _ = sqlDB.Exec(`PRAGMA foreign_keys = ON;`)
		_, _ = sqlDB.Exec(`PRAGMA journal_mode = WAL;`)
		_, _ = sqlDB.Exec(`PRAGMA synchronous = NORMAL;`)
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
	return sqlDB.Ping()
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
