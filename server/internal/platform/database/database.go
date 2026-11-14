package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a pgx pool from a libpq connection string (keyword/value or URL).
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		// Don't wrap err: pgx parse errors can echo the DSN, which contains the password.
		return nil, fmt.Errorf("database: invalid connection config (check DB_* variables)")
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}
	return pool, nil
}
