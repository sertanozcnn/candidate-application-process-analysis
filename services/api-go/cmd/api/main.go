package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"capa/services/api-go/internal/config"
	"capa/services/api-go/internal/modules/health"
	"capa/services/api-go/internal/platform/httpserver"
	"capa/services/api-go/internal/platform/logger"
	"capa/services/api-go/internal/platform/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log := logger.New(config.Config{})
		log.Error().Err(err).Msg("config failed")
		os.Exit(1)
	}

	log := logger.New(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error().Err(err).Msg("database connection failed")
		os.Exit(1)
	}
	defer db.Close()

	mux := http.NewServeMux()
	health.RegisterRoutes(mux, db)
	handler := httpserver.AccessLog(log, mux)

	server := httpserver.New(httpserver.Config{
		Addr:            cfg.HTTPAddr,
		ShutdownTimeout: cfg.ShutdownTimeout,
		Handler:         handler,
		Logger:          &log,
	})

	if err := server.Run(ctx); err != nil {
		log.Error().Err(err).Msg("server stopped")
		os.Exit(1)
	}
}
