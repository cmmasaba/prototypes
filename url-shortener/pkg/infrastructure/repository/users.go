package repository

import (
	"context"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
)

// SaveUser stores user record in the db
func (r *Repository) SaveUser(ctx context.Context, input *domain.User) (*domain.User, error) {
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

// GetUserByEmail get a user by their email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
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

// GetUserByID get a user by their id
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	res, err := r.db.GetUserByID(ctx, id)
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
