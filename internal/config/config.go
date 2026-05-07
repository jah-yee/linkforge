package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPAddr    string
	PostgresDSN string
	RedisAddr   string
	BaseURL     string
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		PostgresDSN: os.Getenv("POSTGRES_DSN"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),
		BaseURL:     getenv("BASE_URL", "http://localhost:8080"),
	}
	if cfg.PostgresDSN == "" {
		return nil, errors.New("POSTGRES_DSN is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
