package repository

import (
	"context"
	"time"

	"github.com/cmmasaba/prototypes/pkg/application/domain"
	"github.com/cmmasaba/prototypes/pkg/infrastructure/repository/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

var errPreparedStmtNotFound = "prepared statement not found: %s"

func stringToPgtypeText(s *string) pgtype.Text {
	if s != nil {
		return pgtype.Text{String: *s, Valid: true}
	}

	return pgtype.Text{}
}

func pgtypeTextToString(s pgtype.Text) *string {
	var str *string

	if s.Valid {
		str = &s.String
	}

	return str
}

func timeToTimestampz(t *time.Time) pgtype.Timestamptz {
	if t != nil {
		return pgtype.Timestamptz{Time: *t, Valid: true}
	}

	return pgtype.Timestamptz{}
}

func timestampzToTime(t pgtype.Timestamptz) *time.Time {
	var time *time.Time

	if t.Valid {
		time = &t.Time
	}

	return time
}

// SaveShortLink saves the created shortlink to the database.
func (r *Repository) SaveShortLink(ctx context.Context, data domain.Link) (*domain.Link, error) {
	row, err := r.db.SaveShortLink(ctx, sqlc.SaveShortLinkParams{
		ShortCode:      data.ShortCode,
		OriginalUrl:    data.OriginalURL,
		OwnershipToken: data.OwnershipToken,
		ExpiresAt:      timeToTimestampz(data.ExpiresAt),
	})
	if err != nil {
		return nil, err
	}

	return &domain.Link{
		ID:             row.ID,
		ShortCode:      row.ShortCode,
		OriginalURL:    row.OriginalUrl,
		OwnershipToken: row.OwnershipToken,
		CreatedAt:      row.CreatedAt.Time,
		ExpiresAt:      timestampzToTime(row.ExpiresAt),
	}, nil
}

// GetLinkByCode queries links by the given short code.
func (r *Repository) GetLinkByCode(ctx context.Context, code string) (*domain.Link, error) {
	link, err := r.db.GetShortLinkByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return &domain.Link{
		ID:             link.ID,
		ShortCode:      link.ShortCode,
		OriginalURL:    link.OriginalUrl,
		OwnershipToken: link.OwnershipToken,
		CreatedAt:      link.CreatedAt.Time,
		ExpiresAt:      timestampzToTime(link.ExpiresAt),
	}, nil
}
