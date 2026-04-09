// Package repository implements database access functionality.
package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
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

var (
	errAcquireDBConnFailed = errors.New("failed to acquire db connection")
	errPingDBFailed        = errors.New("failed to ping db")
	ErrNotFound            = errors.New("record not found")
)

// Repository encapsulates db operations.
type Repository struct {
	db   *sqlc.Queries
	pool *pgxpool.Pool
}

// New returns a *[Repository] built from the passed connection string.
func New(connString string) (*Repository, error) {
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
	dbConfig.ConnConfig.RuntimeParams["search_path"] = helpers.MustGetEnvVar("DATABASE_SCHEMA")

	ctx := context.Background()

	connPool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, err
	}

	connection, err := connPool.Acquire(ctx)
	if err != nil {
		return nil, errAcquireDBConnFailed
	}

	defer connection.Release()

	err = connection.Ping(ctx)
	if err != nil {
		return nil, errPingDBFailed
	}

	r := &Repository{
		db:   sqlc.New(connPool),
		pool: connPool,
	}

	return r, nil
}

// PingDB returns error if the database connection can't be pinged.
func (r *Repository) PingDB(ctx context.Context) bool {
	ctx, span := telemetry.Trace(ctx, packageName, "PingDB")
	defer span.End()

	connection, err := r.pool.Acquire(ctx)
	if err != nil {
		telemetry.RecordError(span, err)
		slog.Error(errAcquireDBConnFailed.Error(), "err", err)

		return false
	}

	defer connection.Release()

	err = connection.Ping(ctx)
	if err != nil {
		telemetry.RecordError(span, err)
		slog.Error(errPingDBFailed.Error(), "err", err)

		return false
	}

	return true
}
