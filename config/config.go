package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	EncryptionKey string
	Port          string
	GinMode       string
	DBType        string
	DBURL         string
}

func Load() (*Config, error) {
	cfg := &Config{
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		Port:          envOrDefault("PORT", "8090"),
		GinMode:       os.Getenv("GIN_MODE"),
		DBType:        envOrDefault("DB_TYPE", "sqlite"),
		DBURL:         os.Getenv("DB_URL"),
	}

	var errs []string
	if cfg.EncryptionKey == "" {
		errs = append(errs, "ENCRYPTION_KEY is required")
	} else if len(cfg.EncryptionKey) != 32 {
		errs = append(errs, "ENCRYPTION_KEY must be exactly 32 bytes, invalid length")
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(errs, ", "))
	}

	return cfg, nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
