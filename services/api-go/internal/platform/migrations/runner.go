package migrations

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const Dir = "migrations"

func Run(databaseURL string, command string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set migration dialect: %w", err)
	}

	if err := goose.Run(command, db, Dir); err != nil {
		return fmt.Errorf("run migration %q: %w", command, err)
	}

	return nil
}
