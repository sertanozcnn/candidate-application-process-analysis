package logger

import (
	"os"
	"strings"
	"time"

	"capa/services/api-go/internal/config"

	"github.com/rs/zerolog"
)

func New(cfg config.Config) zerolog.Logger {
	level := parseLevel(cfg.LogLevel)
	zerolog.SetGlobalLevel(level)

	if cfg.AppEnv == "production" {
		return zerolog.New(os.Stdout).
			With().
			Timestamp().
			Str("service", "capa-api").
			Logger()
	}

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Kitchen,
	}

	return zerolog.New(output).
		With().
		Timestamp().
		Str("service", "capa-api").
		Logger()
}

func parseLevel(value string) zerolog.Level {
	level, err := zerolog.ParseLevel(strings.ToLower(value))
	if err != nil {
		return zerolog.InfoLevel
	}

	return level
}
