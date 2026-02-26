// Package repository implements functionality for interacting with the database.
package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	insertLinksQuery                      = "insertLinksQuery"
	insertClicksQuery                     = "insertClicksQuery"
	searchLinksByCodeQuery                = "searchLinksByCodeQuery"
	searchLinksByExpiresAtQuery           = "searchLinksByExpiresAtQuery"
	searchClicksByLinkIDAndCountryQuery   = "searchClicksByLinkIDAndCountryQuery"
	searchClicksByLinkIDAndClickedAtQuery = "searchClicksByLinkIDAndClickedAtQuery"
)

type Repository struct {
	DB         *sqlx.DB
	statements map[string]*sqlx.NamedStmt
}

// prepare takes a map of query names and queries then
// creetes a map of query names and prepared statements.
func (r *Repository) prepare(queries map[string]string) error {
	for name, query := range queries {
		stmt, err := r.DB.PrepareNamed(query)
		if err != nil {
			return fmt.Errorf("failed to prepare statement '%s': %w", name, err)
		}

		r.statements[name] = stmt
	}

	return nil
}

// databaseQueries builds a map of name:query for used database queries.
//
// Centralizing queries makes them maintainable.
func databaseQueries() map[string]string {
	return map[string]string{
		insertLinksQuery: `INSERT INTO links (short_code, original_url, ownership_token, expires_at) VALUES (:short_code, :original_url, :ownership_token, :expires_at) RETURNING *`,

		searchLinksByCodeQuery: `SELECT * FROM links WHERE short_code = :short_code`,

		searchLinksByExpiresAtQuery: `SELECT * FROM links WHERE expires_at IS NOT NULL`,

		insertClicksQuery: `INSERT INTO clicks (link_id, clicked_at, ip_hash, referrer, user_agent, device_type, browser, os, country, city) VALUES (:link_id, :clicked_at, :ip_hash, :referrer, :user_agent, :device_type, :browser, :os, :country, :city) RETURNING *`,

		searchClicksByLinkIDAndClickedAtQuery: `SELECT * FROM clicks WHERE link_id = :link_id`,

		searchClicksByLinkIDAndCountryQuery: `SELECT * FROM clicks WHERE country = :country`,
	}
}

// NewRepository returns a [Repository] built from the passed connection string.
//
// connString should be a valid Postgres connection string.
func NewRepository(connString string) (*Repository, error) {
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	r := &Repository{
		DB:         db,
		statements: make(map[string]*sqlx.NamedStmt),
	}

	if err := r.prepare(databaseQueries()); err != nil {
		return nil, err
	}

	return r, nil
}
