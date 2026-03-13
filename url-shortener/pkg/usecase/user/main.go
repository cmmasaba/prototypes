// Package user provides usecases related to user management.
package user

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/user"

	bcryptCount = 14
)

var (
	ErrUserWithEmailExists = errors.New("email already taken")
	ErrInvalidAuthMethod   = errors.New("must provide either password or oauth, not both")
	ErrIncompleteOAuth     = errors.New("both oauth_provider and oauth_provider_id are required")
	ErrNoAuthMethod        = errors.New("must provide either password or oauth credentials")
)

type infrastructure interface {
	CreateUser(ctx context.Context, input *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type UsecaseImpl struct {
	infra infrastructure
}

func New(infrastructure infrastructure) *UsecaseImpl {
	return &UsecaseImpl{
		infra: infrastructure,
	}
}

// CreateUser returns *[dto.UserOutput] after saving user to db
func (u *UsecaseImpl) CreateUser(ctx context.Context, input dto.UserInput) (*dto.UserOutput, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CreateUser")
	defer span.End()

	err := helpers.Validate(input)
	if err != nil {
		slog.ErrorContext(ctx, "input validation failed", "err", err)

		return nil, err
	}

	existingUser, err := u.infra.GetUserByEmail(ctx, input.Email)
	if err != nil {
		slog.ErrorContext(ctx, "error checking for duplicate email", "err", err)

		return nil, err
	}

	if existingUser != nil {
		return nil, ErrUserWithEmailExists
	}

	data, err := mapDTOToDomain(&input)
	if err != nil {
		slog.ErrorContext(ctx, "error mapping input data", "err", err)

		return nil, err
	}

	user, err := u.infra.CreateUser(ctx, data)
	if err != nil {
		slog.ErrorContext(ctx, "error creating user in db", "err", err)

		return nil, err
	}

	return &dto.UserOutput{
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func mapDTOToDomain(data *dto.UserInput) (*domain.User, error) {
	hasPassword := data.Password != ""
	hasOAuth := data.OauthProvider != "" || data.OauthProviderID != ""

	if hasPassword && hasOAuth {
		return nil, ErrInvalidAuthMethod
	}

	if !hasPassword && !hasOAuth {
		return nil, ErrNoAuthMethod
	}

	if hasOAuth && (data.OauthProvider == "" || data.OauthProviderID == "") {
		return nil, ErrIncompleteOAuth
	}

	user := domain.User{
		Email: strings.ToLower(strings.TrimSpace(data.Email)),
	}

	if hasOAuth {
		user.OauthProvider = &data.OauthProvider
		user.OauthProviderID = &data.OauthProviderID
	}

	if hasPassword {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcryptCount)
		if err != nil {
			return nil, err
		}

		hashString := string(hashedPassword)
		user.PasswordHash = &hashString
	}

	return &user, nil
}
