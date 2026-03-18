// Package auth provides features for auth and user management.
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/mail"
	"os"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	packageName           = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/user"
	bcryptCount           = 14
	minEntropy            = 60
	lowercaseCharPoolSize = 26
	uppercaseCharPoolSize = 26
	specialCharPoolSize   = 32
	digitsCharPoolSize    = 10
	accessTokenTTL        = 15 * time.Minute
	refreshTokenTTL       = 60 * time.Minute
)

var (
	accessKey  = os.Getenv("JWT_ACCESS_SIGNING_KEY")
	refreshKey = os.Getenv("JWT_REFRESH_SIGNING_KEY")

	ErrUserWithEmailExists = errors.New("email already in use")
	ErrIncompleteOAuth     = errors.New("both oauth_provider and oauth_provider_id are required")
	ErrInvalidEmailSyntax  = errors.New("invalid email address")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	errInternalError       = errors.New("internal server error")
	ErrInvalidToken        = errors.New("invalid token")
	ErrExpiredToken        = errors.New("token expired")
)

type infrastructure interface {
	CreateUser(ctx context.Context, input *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CheckPasswordIsBreached(ctx context.Context, password string) (bool, error)
	GetRefreshTokenByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error)
	SaveRefreshToken(ctx context.Context, input domain.RefreshToken) error
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
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
) (*dto.AuthResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CreateUser")
	defer span.End()

	e, err := mail.ParseAddress(input.Email)
	if err != nil || e.Address != input.Email {
		slog.ErrorContext(ctx, "validation for email address failed", "err", err)
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

	accessToken, refreshToken, err := u.generateAuthTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		User: dto.User{
			PublicID:  user.PublicID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL / time.Second),
	}, nil
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

// verifyPassword returns nil if a password is successfully verified.
func verifyPassword(ctx context.Context, passwordHash, password string) error {
	_, span := telemetry.Trace(ctx, packageName, "verifyPassword")
	defer span.End()

	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
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

// generateAccessToken return a jwt token and nil error on success.
func generateAccessToken(ctx context.Context, user *domain.User) (string, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "generateAccessToken")
	defer span.End()

	now := time.Now()

	accessTokenClaims := jwt.MapClaims{
		"exp": now.Add(accessTokenTTL).Unix(),
		"iat": now.Unix(),
		"sub": user.PublicID,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessTokenClaims)

	accessTokenString, err := accessToken.SignedString([]byte(accessKey))
	if err != nil {
		slog.ErrorContext(ctx, "jwt access token signing failed", "err", err)
		telemetry.RecordError(span, err)

		return "", err
	}

	return accessTokenString, nil
}

// generateRefreshToken return a jwt token and nil error on success.
func generateRefreshToken(ctx context.Context, user *domain.User) (string, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "generateRefreshToken")
	defer span.End()

	now := time.Now()

	refreshTokenClaims := jwt.MapClaims{
		"sub": user.PublicID,
		"exp": now.Add(refreshTokenTTL).Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshTokenClaims)

	refreshTokenString, err := refreshToken.SignedString([]byte(refreshKey))
	if err != nil {
		slog.ErrorContext(ctx, "jwt refresh token signing failed", "err", err)
		telemetry.RecordError(span, err)

		return "", err
	}

	return refreshTokenString, nil
}

// ValidateJWTToken returns the [jwt.MapClaims] and nil error successful after token validation and verification.
func (u *UsecaseImplUser) ValidateJWTToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "ValidateJWTToken")
	defer span.End()

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return accessKey, nil
	})
	if err != nil {
		if errors.Is(err, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}

		slog.ErrorContext(ctx, "validate jwt token failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// generateAuthTokens returns the access and refresh tokens and nil error on success.
func (u *UsecaseImplUser) generateAuthTokens(ctx context.Context, user *domain.User) (string, string, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "generateAuthTokens")
	defer span.End()

	accessToken, err := generateAccessToken(ctx, user)
	if err != nil {
		return "", "", errInternalError
	}

	refreshToken, err := generateRefreshToken(ctx, user)
	if err != nil {
		return "", "", errInternalError
	}

	err = u.infra.SaveRefreshToken(ctx, domain.RefreshToken{
		UserID:   user.ID,
		Token:    refreshToken,
		ExpireAt: time.Now().Add(refreshTokenTTL),
		Revoked:  false,
	})
	if err != nil {
		slog.ErrorContext(ctx, "save refresh token failed", "err", err)

		return "", "", errInternalError
	}

	return accessToken, refreshToken, nil
}

// Login returns *[dto.LoginResponse] and nil error if user authenticates successfully.
func (u *UsecaseImplUser) Login(ctx context.Context, input *dto.LoginInput) (*dto.AuthResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "Login")
	defer span.End()

	user, err := u.infra.GetUserByEmail(ctx, input.Email)
	if err != nil {
		slog.ErrorContext(ctx, "get user by email failed", "err", err)

		return nil, ErrInvalidCredentials
	}

	err = verifyPassword(ctx, *user.PasswordHash, input.Password)
	if err != nil {
		slog.ErrorContext(ctx, "password verification failed")
		telemetry.RecordError(span, err)

		return nil, ErrInvalidCredentials
	}

	accessToken, refreshToken, err := u.generateAuthTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		User: dto.User{
			PublicID:  user.PublicID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL / time.Second),
	}, nil
}

// RefreshAccessToken returns a new access token.
func (u *UsecaseImplUser) RefreshAccessToken(ctx context.Context, refreshToken string) (*dto.RefreshAccessTokenResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "RefreshAccessToken")
	defer span.End()

	token, err := u.infra.GetRefreshTokenByTokenHash(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if token.Revoked {
		return nil, ErrInvalidToken
	}

	if time.Now().After(token.ExpireAt) {
		return nil, ErrExpiredToken
	}

	user, err := u.infra.GetUserByID(ctx, token.UserID)
	if err != nil {
		telemetry.RecordError(span, fmt.Errorf("get user for refresh token failed: %w", err))

		return nil, errInternalError
	}

	accessToken, err := generateAccessToken(ctx, user)
	if err != nil {
		return nil, errInternalError
	}

	return &dto.RefreshAccessTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(accessTokenTTL / time.Second),
	}, nil
}
