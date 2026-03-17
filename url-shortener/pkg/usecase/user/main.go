// Package user provides usecases related to user management.
package user // nolint: revive

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/mail"
	"unicode"
	"unicode/utf8"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"golang.org/x/crypto/bcrypt"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/user"

	bcryptCount           = 14
	minEntropy            = 60
	lowercaseCharPoolSize = 26
	uppercaseCharPoolSize = 26
	specialCharPoolSize   = 32
	digitsCharPoolSize    = 10
)

var (
	ErrUserWithEmailExists = errors.New("email already taken")
	ErrIncompleteOAuth     = errors.New("both oauth_provider and oauth_provider_id are required")
	ErrInvalidEmailSyntax  = errors.New("invalid email address")
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

// hashPassword returns hashed password string and nil if no error occurred.
//
// bcrypt is intentionally slow making this step critical section in the flow.
func hashPassword(ctx context.Context, password string) (string, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "hashPassword")
	defer span.End()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCount)
	if err != nil {
		slog.ErrorContext(ctx, "bcrypt password hashing failed", "err", err)
		telemetry.RecordError(span, err)

		return "", err
	}

	return string(hashedPassword), nil
}

// ValidatePasswordStrength returns true if the password entropy is >= than minEntropy.
func (u *UsecaseImplUser) ValidatePasswordStrength(ctx context.Context, input *dto.ValidatePasswordInput) bool {
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
func (u *UsecaseImplUser) CheckPasswordIsBreached(ctx context.Context, input *dto.ValidatePasswordInput) bool {
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
	input *dto.EmailPasswordUserInput,
) (*dto.UserOutput, error) {
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

	return &dto.UserOutput{
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}
