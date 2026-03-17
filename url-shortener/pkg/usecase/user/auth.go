package user

import (
	"context"
	"errors"
	"log/slog"
	"math"
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
	accessKey  = os.Getenv("JWT_SIGNING_KEY")
	refreshKey = os.Getenv("JWT_REFRESH_KEY")
)

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

// generateAccessAndRefreshToken return a jwt token and nil error on success.
func generateAccessAndRefreshToken(ctx context.Context, user *domain.User) (string, string, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "generateAccessAndRefreshToken")
	defer span.End()

	now := time.Now()

	accessTokenClaims := jwt.MapClaims{
		"exp": now.Add(accessTokenTTL),
		"iat": now.Unix(),
		"sub": user.Email,
	}

	refreshTokenClaims := jwt.MapClaims{
		"sub": user.Email,
		"exp": now.Add(refreshTokenTTL),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessTokenClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshTokenClaims)

	accessTokenString, err := accessToken.SignedString(accessKey)
	if err != nil {
		slog.ErrorContext(ctx, "jwt access token signing failed", "err", err)
		telemetry.RecordError(span, err)

		return "", "", err
	}

	refreshTokenString, err := refreshToken.SignedString(refreshKey)
	if err != nil {
		slog.ErrorContext(ctx, "jwt refresh token signing failed", "err", err)
		telemetry.RecordError(span, err)

		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// validateJWTToken returns the [jwt.MapClaims] and nil error successful after token validation and verification.
func validateJWTToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "validateJWTToken")
	defer span.End()

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errInvalidToken
		}

		return accessKey, nil
	})
	if err != nil {
		if errors.Is(err, errExpiredToken) {
			return nil, errExpiredToken
		}

		slog.ErrorContext(ctx, "validate jwt token failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, errInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errInvalidToken
}

// Login returns *[dto.LoginResponse] and nil error if user authenticates successfully.
func (u *UsecaseImplUser) Login(ctx context.Context, input *dto.LoginInput) (*dto.LoginResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "Login")
	defer span.End()

	user, err := u.infra.GetUserByEmail(ctx, input.Email)
	if err != nil {
		slog.ErrorContext(ctx, "get user by email failed", "err", err)

		return nil, errInvalidCredentials
	}

	err = verifyPassword(ctx, *user.PasswordHash, input.Password)
	if err != nil {
		slog.ErrorContext(ctx, "password verification failed")
		telemetry.RecordError(span, err)

		return nil, errInvalidCredentials
	}

	accessToken, refreshToken, err := generateAccessAndRefreshToken(ctx, user)
	if err != nil {
		return nil, errInternalError
	}

	err = u.infra.SaveRefreshToken(ctx, domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshToken,
		ExpireAt:  time.Now().Add(refreshTokenTTL),
		Revoked:   false,
	})
	if err != nil {
		slog.ErrorContext(ctx, "save refresh token failed", "err", err)

		return nil, errInvalidCredentials
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
