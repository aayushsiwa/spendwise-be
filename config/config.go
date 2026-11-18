package config

import (
	"os"
)

type Config struct {
	Port    string
	GinMode string
	DBType  string
	DBURL   string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:    envOrDefault("PORT", "8090"),
		GinMode: os.Getenv("GIN_MODE"),
		DBType:  envOrDefault("DB_TYPE", "sqlite"),
		DBURL:   os.Getenv("DB_URL"),
	}

	return cfg, nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
