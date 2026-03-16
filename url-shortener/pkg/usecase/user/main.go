// Package user provides usecases related to user management.
package user // nolint: revive

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/user"

	bcryptCount = 14

	minEntropy            = 60
	lowercaseCharPoolSize = 26
	uppercaseCharPoolSize = 26
	specialCharPoolSize   = 32
	digitsCharPoolSize    = 10
)

var (
	ErrUserWithEmailExists = errors.New("email already taken")
	ErrIncompleteOAuth     = errors.New("both oauth_provider and oauth_provider_id are required")
)

type infrastructure interface {
	CreateUser(ctx context.Context, input *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CheckPasswordIsBreached(ctx context.Context, password string) (bool, error)
}

type UsecaseImplUser struct {
	infra infrastructure
}

func New(infrastructure infrastructure) *UsecaseImplUser {
	return &UsecaseImplUser{
		infra: infrastructure,
	}
}

// getBase counts the size of the character pool used.
func getBase(password string) int {
	var (
		base                                     int
		hasLower, hasUpper, hasSpecial, hasDigit bool
	)

	for _, ch := range password {
		switch {
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasDigit = true
		default:
			hasSpecial = true
		}

		if hasLower && hasUpper && hasSpecial && hasDigit {
			break
		}
	}

	if hasDigit {
		base += digitsCharPoolSize
	}

	if hasSpecial {
		base += specialCharPoolSize
	}

	if hasLower {
		base += lowercaseCharPoolSize
	}

	if hasUpper {
		base += uppercaseCharPoolSize
	}

	return base
}

// ValidatePasswordStrength returns true if the password entropy is >= than minEntropy.
func (u *UsecaseImplUser) ValidatePasswordStrength(ctx context.Context, input dto.ValidatePasswordInput) bool {
	_, span := telemetry.Trace(ctx, packageName, "ValidatePasswordStrength")
	defer span.End()

	base := getBase(input.Password)

	// get number of characters used. len(password) would incorrectly return byte count.
	length := utf8.RuneCountInString(input.Password)

	// Entropy = log2(base^length)
	entropy := float64(length) * math.Log2(float64(base))

	return entropy >= minEntropy
}

// CheckPasswordIsBreached returns true is a password was found in a breach.
func (u *UsecaseImplUser) CheckPasswordIsBreached(ctx context.Context, input dto.ValidatePasswordInput) bool {
	ctx, span := telemetry.Trace(ctx, packageName, "CheckPasswordIsBreached")
	defer span.End()

	isBreached, err := u.infra.CheckPasswordIsBreached(ctx, input.Password)
	if err != nil {
		telemetry.RecordError(span, err)

		return isBreached
	}

	return isBreached
}

// CreateUserEmailPassword returns *[dto.UserOutput] after saving user to db
func (u *UsecaseImplUser) CreateUserEmailPassword(
	ctx context.Context,
	input dto.EmailPasswordUserInput,
) (*dto.UserOutput, error) {
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCount)
	if err != nil {
		return nil, err
	}

	hashString := string(hashedPassword)

	user, err := u.infra.CreateUser(ctx, &domain.User{
		Email:        input.Email,
		PasswordHash: &hashString,
	})
	if err != nil {
		slog.ErrorContext(ctx, "error creating user in db", "err", err)

		return nil, err
	}

	return &dto.UserOutput{
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}
