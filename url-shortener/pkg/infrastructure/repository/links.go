package repository

import (
	"context"
	"errors"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
	"github.com/google/uuid"
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

func timeToTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t != nil {
		return pgtype.Timestamptz{Time: *t, Valid: true}
	}

	return pgtype.Timestamptz{}
}

func timestamptzToTime(t pgtype.Timestamptz) *time.Time {
	var time *time.Time

	if t.Valid {
		time = &t.Time
	}

	return time
}

func pgtypeUUIDToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}

	return uuid.UUID(u.Bytes).String()
}

func stringToPgtypeUUID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

// CreateShortLink returns a *[domain.Link] created from the input data.
func (r *Repository) CreateShortLink(ctx context.Context, data domain.Link) (*domain.Link, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CreateShortLink")
	defer span.End()

	res, err := r.db.SaveShortLink(ctx, sqlc.SaveShortLinkParams{
		UserID:         data.UserID,
		ShortCode:      data.ShortCode,
		OriginalUrl:    data.OriginalURL,
		OwnershipToken: data.OwnershipToken,
		ExpiresAt:      timeToTimestamptz(data.ExpiresAt),
	})
	if err != nil {
		telemetry.RecordError(span, err)

		return nil, err
	}

	return &domain.Link{
		ID:             res.ID,
		UserID:         res.UserID,
		ShortCode:      res.ShortCode,
		OriginalURL:    res.OriginalUrl,
		OwnershipToken: res.OwnershipToken,
	}, nil
}

// GetLinkByCode returns a *[domain.Link] matching the short code.
func (r *Repository) GetLinkByCode(ctx context.Context, code string) (*domain.Link, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetLinkByCode")
	defer span.End()

	res, err := r.db.GetShortLinkByCode(ctx, code)
	if err != nil {
		telemetry.RecordError(span, err)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &domain.Link{
		ID:             res.ID,
		UserID:         res.UserID,
		ShortCode:      res.ShortCode,
		OriginalURL:    res.OriginalUrl,
		OwnershipToken: res.OwnershipToken,
	}, nil
}
