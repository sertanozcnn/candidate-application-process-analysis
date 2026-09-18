package main

import (
	"fmt"
	"os"

	"capa/services/api-go/internal/config"
	"capa/services/api-go/internal/platform/migrations"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	return migrations.Run(cfg.DatabaseURL, command)
}
