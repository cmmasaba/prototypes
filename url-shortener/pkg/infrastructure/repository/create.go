package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cmmasaba/prototypes/pkg/application/domain"
)

var (
	errPreparedStmtNotFound = "prepared statement not found: %s"
	errReadingRowIntoStruct = "failed to read row into struct: %w"
)

// SaveShortLink saves the created shortlink to the database.
func (r *Repository) SaveShortLink(ctx context.Context, data domain.Link) (*domain.Link, error) {
	link := Link{
		ShortCode:      data.ShortCode,
		OriginalURL:    data.OriginalURL,
		OwnershipToken: data.OwnershipToken,
		ExpiresAt:      timeToNullTime(data.ExpiresAt),
	}

	stmt, ok := r.statements[insertLinksQuery]
	if !ok {
		return nil, fmt.Errorf(errPreparedStmtNotFound, insertLinksQuery)
	}

	row := stmt.QueryRowxContext(ctx, link)

	err := row.StructScan(&link)
	if err != nil {
		return nil, fmt.Errorf(errReadingRowIntoStruct, err)
	}

	return &domain.Link{
		ID:             link.ID,
		ShortCode:      link.ShortCode,
		OriginalURL:    link.OriginalURL,
		OwnershipToken: link.OwnershipToken,
		CreatedAt:      link.CreatedAt,
		ExpiresAt:      nullTimeToTime(link.ExpiresAt),
	}, nil
}

// SaveClickData saves the click data for a link to the database.
func (r *Repository) SaveClickData(ctx context.Context, data domain.Click) (*domain.Click, error) {
	click := Click{
		ClickedAt:  data.ClickedAt,
		LinkID:     data.LinkID,
		IPHash:     stringToNullString(data.IPHash),
		Referrer:   stringToNullString(data.Referrer),
		UserAgent:  stringToNullString(data.UserAgent),
		DeviceType: stringToNullString(data.DeviceType),
		Browser:    stringToNullString(data.Browser),
		OS:         stringToNullString(data.OS),
		Country:    stringToNullString(data.Country),
		City:       stringToNullString(data.City),
	}

	stmt, ok := r.statements[insertClicksQuery]
	if !ok {
		return nil, fmt.Errorf(errPreparedStmtNotFound, insertClicksQuery)
	}

	row := stmt.QueryRowxContext(ctx, click)

	err := row.StructScan(&click)
	if err != nil {
		return nil, fmt.Errorf(errReadingRowIntoStruct, err)
	}

	return &domain.Click{
		ID:         click.ID,
		LinkID:     click.LinkID,
		ClickedAt:  click.ClickedAt,
		IPHash:     nullStringToString(click.IPHash),
		Referrer:   nullStringToString(click.Referrer),
		UserAgent:  nullStringToString(click.UserAgent),
		DeviceType: nullStringToString(click.DeviceType),
		Browser:    nullStringToString(click.Browser),
		OS:         nullStringToString(click.OS),
		Country:    nullStringToString(click.Country),
		City:       nullStringToString(click.City),
	}, nil
}

func stringToNullString(s *string) sql.NullString {
	if s != nil {
		return sql.NullString{String: *s, Valid: true}
	}

	return sql.NullString{}
}

func nullStringToString(s sql.NullString) *string {
	var str *string

	if s.Valid {
		str = &s.String
	}

	return str
}

func timeToNullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}

	return sql.NullTime{}
}

func nullTimeToTime(t sql.NullTime) *time.Time {
	var time *time.Time

	if t.Valid {
		time = &t.Time
	}

	return time
}
