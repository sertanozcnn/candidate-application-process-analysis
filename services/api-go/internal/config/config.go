package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultAppEnv           = "development"
	defaultLogLevel         = "info"
	defaultHTTPAddr         = ":8080"
	defaultDatabaseURL      = "postgres://capa:capa_local_password@localhost:55432/capa?sslmode=disable"
	defaultShutdownTimeout  = 10 * time.Second
	defaultAdminOrigins     = "http://localhost:3001,http://127.0.0.1:3001"
	defaultCandidateOrigins = "http://localhost:3000,http://127.0.0.1:3000"
)

type Config struct {
	AppEnv           string
	LogLevel         string
	HTTPAddr         string
	DatabaseURL      string
	AdminOrigins     []string
	CandidateOrigins []string
	ShutdownTimeout  time.Duration
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppEnv:           env("CAPA_APP_ENV", defaultAppEnv),
		LogLevel:         env("CAPA_LOG_LEVEL", defaultLogLevel),
		HTTPAddr:         env("CAPA_API_ADDR", defaultHTTPAddr),
		DatabaseURL:      env("CAPA_DATABASE_URL", defaultDatabaseURL),
		AdminOrigins:     splitList(env("CAPA_ADMIN_ORIGINS", defaultAdminOrigins)),
		CandidateOrigins: splitList(env("CAPA_CANDIDATE_ORIGINS", defaultCandidateOrigins)),
		ShutdownTimeout:  defaultShutdownTimeout,
	}

	if value := os.Getenv("CAPA_API_SHUTDOWN_TIMEOUT"); value != "" {
		timeout, err := time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("invalid CAPA_API_SHUTDOWN_TIMEOUT: %w", err)
		}
		cfg.ShutdownTimeout = timeout
	}

	if cfg.AppEnv == "" {
		return Config{}, fmt.Errorf("CAPA_APP_ENV must not be empty")
	}
	if cfg.LogLevel == "" {
		return Config{}, fmt.Errorf("CAPA_LOG_LEVEL must not be empty")
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("CAPA_API_ADDR must not be empty")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("CAPA_DATABASE_URL must not be empty")
	}
	if len(cfg.AdminOrigins) == 0 {
		return Config{}, fmt.Errorf("CAPA_ADMIN_ORIGINS must not be empty")
	}
	if len(cfg.CandidateOrigins) == 0 {
		return Config{}, fmt.Errorf("CAPA_CANDIDATE_ORIGINS must not be empty")
	}
	if cfg.ShutdownTimeout <= 0 {
		return Config{}, fmt.Errorf("CAPA_API_SHUTDOWN_TIMEOUT must be positive")
	}

	return cfg, nil
}

func splitList(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			values = append(values, item)
		}
	}
	return values
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
