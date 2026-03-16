// Package repository implements functionality for interacting with the database.
package repository

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns          = int32(4)
	defaultMinConns          = int32(0)
	defaultMaxConnLifetime   = time.Hour
	defaultMaxConnIdleTime   = time.Minute * 30
	defaultHealthCheckPeriod = time.Minute
	defaultConnectTimeout    = time.Second * 5

	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
)

// Repository encapsulates db operations.
type Repository struct {
	db   *sqlc.Queries
	pool *pgxpool.Pool
}

// New returns a [Repository] built from the passed connection string.
//
// connString should be a valid Postgres connection string.
func New() (*Repository, error) {
	connString, ok := os.LookupEnv("POSTGRES_URL")
	if !ok {
		return nil, fmt.Errorf("database connection string string not set")
	}

	dbConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgxpool config: %w", err)
	}

	dbConfig.MaxConns = defaultMaxConns
	dbConfig.MinConns = defaultMinConns
	dbConfig.MaxConnLifetime = defaultMaxConnLifetime
	dbConfig.MaxConnIdleTime = defaultMaxConnIdleTime
	dbConfig.HealthCheckPeriod = defaultHealthCheckPeriod
	dbConfig.ConnConfig.ConnectTimeout = defaultConnectTimeout

	ctx := context.Background()

	connPool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, err
	}

	connection, err := connPool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire pool connection: %w", err)
	}

	defer connection.Release()

	err = connection.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	r := &Repository{
		db:   sqlc.New(connPool),
		pool: connPool,
	}

	return r, nil
}

// PingDB returns error if the database connection can't be pinged.
func (r *Repository) PingDB(ctx context.Context) error {
	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		slog.Error("failed to acquire pool connection", "err", err)

		return fmt.Errorf("failed to acquire pool connection: %w", err)
	}

	defer connection.Release()

	err = connection.Ping(ctx)
	if err != nil {
		slog.Error("failed to ping database", "err", err)

		return fmt.Errorf("failed to ping db: %w", err)
	}

	return nil
}
