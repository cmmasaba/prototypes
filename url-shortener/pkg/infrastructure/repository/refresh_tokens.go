package repository

import (
	"context"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
)

// SaveRefreshToken stores a refresh token in the db
func (r *Repository) SaveRefreshToken(ctx context.Context, input domain.RefreshToken) error {
	err := r.db.SaveRefreshToken(ctx, sqlc.SaveRefreshTokenParams{
		UserID:    input.UserID,
		TokenHash: input.TokenHash,
		ExpiresAt: timeToTimestampz(&input.ExpireAt),
	})
	if err != nil {
		return err
	}

	return nil
}
