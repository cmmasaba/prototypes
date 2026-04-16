package repository

import (
	"context"
	"errors"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

// GetLoginAttempt returns the *[domain.LoginAttempt] for userID given.
func (r *Repository) GetLoginAttempt(ctx context.Context, userID int64) (*domain.LoginAttempt, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetLoginAttempt")
	defer span.End()

	res, err := r.db.GetLoginAttempt(ctx, userID)
	if err != nil {
		telemetry.RecordError(span, err)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &domain.LoginAttempt{
		FailCount:   int(res.FailCount),
		Tier:        int(res.Tier),
		LockedUntil: timestamptzToTime(res.LockedUntil),
		UpdatedAt:   timestamptzToTime(res.UpdatedAt),
	}, nil
}

// UpdateLoginAttempt inserts or updates a login attempt tied to the userID
func (r *Repository) UpdateLoginAttempt(ctx context.Context, userID int64, failCount, tier int, lockedUntil *time.Time) error {
	ctx, span := telemetry.Trace(ctx, packageName, "UpsertLoginAttempt")
	defer span.End()

	err := r.db.UpsertLoginAttempt(ctx, sqlc.UpsertLoginAttemptParams{
		FailCount:   int32(failCount), // nolint: gosec
		Tier:        int32(tier),      // nolint: gosec
		LockedUntil: timeToTimestamptz(lockedUntil),
		UserID:      userID,
	})
	if err != nil {
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}

// ResetLoginAttempts resets the number of login attempts tied to userID.
func (r *Repository) ResetLoginAttempts(ctx context.Context, userID int64) error {
	ctx, span := telemetry.Trace(ctx, packageName, "ResetLoginAttempts")
	defer span.End()

	err := r.db.ResetLoginAttempts(ctx, userID)
	if err != nil {
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}
