package repository

import (
	"database/sql"
	"time"
)

type Base struct {
	ID        int64     `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

type Link struct {
	Base

	ShortCode      string       `db:"short_code"`
	OriginalURL    string       `db:"original_url"`
	OwnershipToken string       `db:"ownership_token"`
	ExpiresAt      sql.NullTime `db:"expires_at"`
}

type Click struct {
	Base

	LinkID     int64          `db:"link_id"`
	ClickedAt  time.Time      `db:"clicked_at"`
	IPHash     sql.NullString `db:"ip_hash"`
	Referrer   sql.NullString `db:"referrer"`
	UserAgent  sql.NullString `db:"user_agent"`
	DeviceType sql.NullString `db:"device_type"`
	Browser    sql.NullString `db:"browser"`
	OS         sql.NullString `db:"os"`
	Country    sql.NullString `db:"country"`
	City       sql.NullString `db:"city"`
}
