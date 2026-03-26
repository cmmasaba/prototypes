package repository

import (
	"context"
	"errors"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

// CreateUser returns a *[domain.User] created from the input data.
func (r *Repository) CreateUser(ctx context.Context, input *domain.User) (*domain.User, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CreateUser")
	defer span.End()

	res, err := r.db.SaveUser(ctx, sqlc.SaveUserParams{
		Email:           input.Email,
		Password:        stringToPgtypeText(input.PasswordHash),
		OauthProvider:   stringToPgtypeText(input.OauthProvider),
		OauthProviderID: stringToPgtypeText(input.OauthProviderID),
	})
	if err != nil {
		telemetry.RecordError(span, err)

		return nil, err
	}

	input.ID = res.ID
	input.PublicID = pgtypeUUIDToString(res.PublicID)
	input.CreatedAt = res.CreatedAt.Time

	return input, nil
}

// GetUserByEmail returns a *[domain.User] matching the email.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetUserByEmail")
	defer span.End()

	res, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		telemetry.RecordError(span, err)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &domain.User{
		ID:              res.ID,
		PublicID:        pgtypeUUIDToString(res.PublicID),
		Email:           res.Email,
		PasswordHash:    pgtypeTextToString(res.Password),
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
		telemetry.RecordError(span, err)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &domain.User{
		ID:              res.ID,
		PublicID:        pgtypeUUIDToString(res.PublicID),
		Email:           res.Email,
		PasswordHash:    pgtypeTextToString(res.Password),
		OauthProvider:   pgtypeTextToString(res.OauthProvider),
		OauthProviderID: pgtypeTextToString(res.OauthProviderID),
		CreatedAt:       res.CreatedAt.Time,
	}, nil
}

// GetUserByID returns a *[domain.User] matching the id.
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetUserByID")
	defer span.End()

	res, err := r.db.GetUserByID(ctx, id)
	if err != nil {
		telemetry.RecordError(span, err)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &domain.User{
		ID:              res.ID,
		PublicID:        pgtypeUUIDToString(res.PublicID),
		Email:           res.Email,
		PasswordHash:    pgtypeTextToString(res.Password),
		OauthProvider:   pgtypeTextToString(res.OauthProvider),
		OauthProviderID: pgtypeTextToString(res.OauthProviderID),
		CreatedAt:       res.CreatedAt.Time,
	}, nil
}

// GetUserByPublicID returns a *[domain.User] matching the public ID.
func (r *Repository) GetUserByPublicID(ctx context.Context, publicID string) (*domain.User, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetUserByPublicID")
	defer span.End()

	pgUUID, err := stringToPgtypeUUID(publicID)
	if err != nil {
		telemetry.RecordError(span, err)

		return nil, err
	}

	res, err := r.db.GetUserByPublicID(ctx, pgUUID)
	if err != nil {
		telemetry.RecordError(span, err)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &domain.User{
		ID:              res.ID,
		PublicID:        pgtypeUUIDToString(res.PublicID),
		Email:           res.Email,
		PasswordHash:    pgtypeTextToString(res.Password),
		OauthProvider:   pgtypeTextToString(res.OauthProvider),
		OauthProviderID: pgtypeTextToString(res.OauthProviderID),
		CreatedAt:       res.CreatedAt.Time,
	}, nil
}
