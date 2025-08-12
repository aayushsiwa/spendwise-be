package db

import (
	"database/sql"
	"log/slog"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Init initializes the database connection with proper error handling
func Init(path string) error {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		slog.Error("Failed to open database", "path", path, "error", err)
		return err
	}

	if err = DB.Ping(); err != nil {
		slog.Error("Database connection failed", "path", path, "error", err)
		return err
	}

	// Enable foreign key constraints
	_, err = DB.Exec(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		slog.Error("Failed to enable foreign keys", "error", err)
		return err
	}

	// Set WAL mode for better concurrency
	_, err = DB.Exec(`PRAGMA journal_mode = WAL;`)
	if err != nil {
		slog.Error("Failed to set WAL mode", "error", err)
		return err
	}

	// Set synchronous mode for better performance vs safety trade-off
	_, err = DB.Exec(`PRAGMA synchronous = NORMAL;`)
	if err != nil {
		slog.Error("Failed to set synchronous mode", "error", err)
		return err
	}

	slog.Info("Database connected successfully", "path", path)
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		slog.Info("Closing database connection")
		return DB.Close()
	}
	return nil
}

// HealthCheck performs a database health check
func HealthCheck() error {
	if DB == nil {
		return sql.ErrConnDone
	}
	return DB.Ping()
}

// GetStats returns database statistics
func GetStats() sql.DBStats {
	if DB == nil {
		return sql.DBStats{}
	}
	return DB.Stats()
}
