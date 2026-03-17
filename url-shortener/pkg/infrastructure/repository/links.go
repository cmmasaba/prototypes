package repository

import (
	"context"
	"errors"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

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

// CreateShortLink returns a *[domain.Link] created from the input data.
func (r *Repository) CreateShortLink(ctx context.Context, data domain.Link) (*domain.Link, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CreateShortLink")
	defer span.End()

	row, err := r.db.SaveShortLink(ctx, sqlc.SaveShortLinkParams{
		ShortCode:      data.ShortCode,
		OriginalUrl:    data.OriginalURL,
		OwnershipToken: data.OwnershipToken,
		ExpiresAt:      timeToTimestampz(data.ExpiresAt),
	})
	if err != nil {
		telemetry.RecordError(span, err)

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

// GetLinkByCode returns a *[domain.Link] matching the short code.
func (r *Repository) GetLinkByCode(ctx context.Context, code string) (*domain.Link, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetLinkByCode")
	defer span.End()

	link, err := r.db.GetShortLinkByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		telemetry.RecordError(span, err)

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
