package repository

import (
	"context"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
)

// CreateUser returns a *[domain.User] created from the input data.
func (r *Repository) CreateUser(ctx context.Context, input *domain.User) (*domain.User, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CreateUser")
	defer span.End()

	res, err := r.db.SaveUser(ctx, sqlc.SaveUserParams{
		Email:           input.Email,
		PasswordHash:    stringToPgtypeText(input.PasswordHash),
		OauthProvider:   stringToPgtypeText(input.OauthProvider),
		OauthProviderID: stringToPgtypeText(input.OauthProviderID),
	})
	if err != nil {
		return nil, err
	}

	input.ID = res.ID
	input.CreatedAt = res.CreatedAt.Time

	return input, nil
}

// GetUserByEmail returns a *[domain.User] matching the email.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetUserByEmail")
	defer span.End()

	res, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:              res.ID,
		Email:           res.Email,
		PasswordHash:    pgtypeTextToString(res.PasswordHash),
		OauthProvider:   pgtypeTextToString(res.OauthProvider),
		OauthProviderID: pgtypeTextToString(res.OauthProviderID),
		CreatedAt:       res.CreatedAt.Time,
	}, nil
}

// GetUserByOAuthID returns a *[domain.User] matching the OAuth provider and provider ID.
func (r *Repository) GetUserByOAuthID(
	ctx context.Context,
	oauthProvider, oauthProviderID string,
) (*domain.User, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetUserByOAuthID")
	defer span.End()

	res, err := r.db.GetUserByOauthID(ctx, sqlc.GetUserByOauthIDParams{
		OauthProvider:   stringToPgtypeText(&oauthProvider),
		OauthProviderID: stringToPgtypeText(&oauthProviderID),
	})
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:              res.ID,
		Email:           res.Email,
		PasswordHash:    pgtypeTextToString(res.PasswordHash),
		OauthProvider:   pgtypeTextToString(res.OauthProvider),
		OauthProviderID: pgtypeTextToString(res.OauthProviderID),
		CreatedAt:       res.CreatedAt.Time,
	}, nil
}
