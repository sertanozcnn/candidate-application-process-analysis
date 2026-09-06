package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const connectTimeout = 5 * time.Second

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, config)
	if err != nil {
		return nil, fmt.Errorf("open postgres pool: %w", err)
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
