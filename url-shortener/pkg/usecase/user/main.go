// Package user provides usecases related to user management.
package user // nolint: revive

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/user"
)

var (
	ErrUserWithEmailExists = errors.New("email already in use")
	ErrIncompleteOAuth     = errors.New("both oauth_provider and oauth_provider_id are required")
	ErrInvalidEmailSyntax  = errors.New("invalid email address")
	errInvalidCredentials  = errors.New("invalid credentials")
	errInternalError       = errors.New("internal server error")
	errInvalidToken        = errors.New("invalid token")
	errExpiredToken        = errors.New("token expired")
)

type infrastructure interface {
	CreateUser(ctx context.Context, input *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CheckPasswordIsBreached(ctx context.Context, password string) (bool, error)
	GetRefreshTokenByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error)
	SaveRefreshToken(ctx context.Context, input domain.RefreshToken) error
}

type UsecaseImplUser struct {
	infra infrastructure
}

func New(infrastructure infrastructure) *UsecaseImplUser {
	return &UsecaseImplUser{
		infra: infrastructure,
	}
}

// CreateUserEmailPassword returns *[dto.UserOutput] after saving user to db
func (u *UsecaseImplUser) CreateUserEmailPassword(
	ctx context.Context,
	input *dto.EmailPasswordUserInput,
) (*dto.User, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CreateUser")
	defer span.End()

	e, err := mail.ParseAddress(input.Email)
	if err != nil || e.Address != input.Email {
		slog.Error("validation for email address failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, ErrInvalidEmailSyntax
	}

	existingUser, err := u.infra.GetUserByEmail(ctx, input.Email)
	if err != nil {
		slog.ErrorContext(ctx, "error checking for duplicate email", "err", err)

		return nil, err
	}

	if existingUser != nil {
		return nil, ErrUserWithEmailExists
	}

	hashString, err := hashPassword(ctx, input.Password)
	if err != nil {
		return nil, err
	}

	user, err := u.infra.CreateUser(ctx, &domain.User{
		Email:        input.Email,
		PasswordHash: &hashString,
	})
	if err != nil {
		slog.ErrorContext(ctx, "error creating user in db", "err", err)

		return nil, err
	}

	return &dto.User{
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}
