// Package repository implements functionality for interacting with the database.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/cmmasaba/prototypes/pkg/infrastructure/repository/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	insertLinksQuery                      = "insertLinksQuery"
	insertClicksQuery                     = "insertClicksQuery"
	searchLinksByCodeQuery                = "searchLinksByCodeQuery"
	searchLinksByExpiresAtQuery           = "searchLinksByExpiresAtQuery"
	searchClicksByLinkIDAndCountryQuery   = "searchClicksByLinkIDAndCountryQuery"
	searchClicksByLinkIDAndClickedAtQuery = "searchClicksByLinkIDAndClickedAtQuery"

	defaultMaxConns          = int32(4)
	defaultMinConns          = int32(0)
	defaultMaxConnLifetime   = time.Hour
	defaultMaxConnIdleTime   = time.Minute * 30
	defaultHealthCheckPeriod = time.Minute
	defaultConnectTimeout    = time.Second * 5
)

type Repository struct {
	statements map[string]string
	db         *sqlc.Queries
}

// databaseQueries builds a map of name:query for used database queries.
//
// Centralizing queries makes them maintainable.
func databaseQueries() map[string]string {
	return map[string]string{
		insertLinksQuery: `INSERT INTO urlshortener.links (short_code, original_url, ownership_token, expires_at) VALUES (:short_code, :original_url, :ownership_token, :expires_at) RETURNING *`,

		searchLinksByCodeQuery: `SELECT * FROM urlshortener.links WHERE short_code = :short_code`,

		searchLinksByExpiresAtQuery: `SELECT * FROM urlshortener.links WHERE expires_at IS NOT NULL`,

		insertClicksQuery: `INSERT INTO urlshortener.clicks (link_id, clicked_at, ip_hash, referrer, user_agent, device_type, browser, os, country, city) VALUES (:link_id, :clicked_at, :ip_hash, :referrer, :user_agent, :device_type, :browser, :os, :country, :city) RETURNING *`,

		searchClicksByLinkIDAndClickedAtQuery: `SELECT * FROM urlshortener.clicks WHERE link_id = :link_id`,

		searchClicksByLinkIDAndCountryQuery: `SELECT * FROM urlshortener.clicks WHERE country = :country`,
	}
}

// NewRepository returns a [Repository] built from the passed connection string.
//
// connString should be a valid Postgres connection string.
func NewRepository(connString string) (*Repository, error) {
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
		return nil, fmt.Errorf("error acquire db pool connection: %w", err)
	}

	defer connection.Release()

	err = connection.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not db: %w", err)
	}

	r := &Repository{
		statements: databaseQueries(),
		db:         sqlc.New(connPool),
	}

	return r, nil
}
