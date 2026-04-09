package repository

import (
	"context"
	"errors"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

// SaveRefreshToken returns nil if refresh token is saved successfully.
func (r *Repository) SaveRefreshToken(ctx context.Context, input domain.RefreshToken) error {
	ctx, span := telemetry.Trace(ctx, packageName, "SaveRefreshToken")
	defer span.End()

	err := r.db.SaveRefreshToken(ctx, sqlc.SaveRefreshTokenParams{
		UserID:    input.UserID,
		Token:     input.Token,
		ExpiresAt: timeToTimestamptz(&input.ExpiresAt),
		Revoked:   input.Revoked,
		CreatedAt: timeToTimestamptz(&input.CreatedAt),
	})
	if err != nil {
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}

// GetRefreshTokenByTokenHash returns a refresh token matching the token hash.
func (r *Repository) GetRefreshTokenByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetRefreshTokenByTokenHash")
	defer span.End()

	res, err := r.db.GetRefreshTokenByToken(ctx, token)
	if err != nil {
		telemetry.RecordError(span, err)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		return nil, err
	}

	return &domain.RefreshToken{
		ID:        res.ID,
		UserID:    res.UserID,
		Token:     res.Token,
		ExpiresAt: *timestamptzToTime(res.ExpiresAt),
		CreatedAt: *timestamptzToTime(res.CreatedAt),
		Revoked:   res.Revoked,
	}, nil
}

// RevokeRefreshToken returns nil on success.
func (r *Repository) RevokeRefreshToken(ctx context.Context, token string) error {
	ctx, span := telemetry.Trace(ctx, packageName, "RevokeRefreshToken")
	defer span.End()

	err := r.db.RevokeRefreshToken(ctx, token)
	if err != nil {
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}

// RevokeAllRefreshTokensForUser returns nil on success.
func (r *Repository) RevokeAllRefreshTokensForUser(ctx context.Context, userID int64) error {
	ctx, span := telemetry.Trace(ctx, packageName, "RevokeAllRefreshTokensForUser")
	defer span.End()

	err := r.db.RevokeAllRefreshTokensForUser(ctx, int64(userID))
	if err != nil {
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}
