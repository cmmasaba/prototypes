package repository

import (
	"context"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
)

// SaveRefreshToken returns nil if refresh token is saved successfully.
func (r *Repository) SaveRefreshToken(ctx context.Context, input domain.RefreshToken) error {
	ctx, span := telemetry.Trace(ctx, packageName, "SaveRefreshToken")
	defer span.End()

	err := r.db.SaveRefreshToken(ctx, sqlc.SaveRefreshTokenParams{
		UserID:    input.UserID,
		TokenHash: input.TokenHash,
		ExpiresAt: timeToTimestampz(&input.ExpireAt),
	})
	if err != nil {
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}
