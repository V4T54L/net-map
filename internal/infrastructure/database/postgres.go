package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates and returns a new PostgreSQL connection pool.
// It pings the database to ensure a connection is established.
func NewPostgresPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

