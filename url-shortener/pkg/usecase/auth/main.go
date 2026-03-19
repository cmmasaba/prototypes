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
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
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
	ErrInvalidOTPCode      = errors.New("invalid otp code, request a new code")
	ErrExpiredOTPCode      = errors.New("expired otp code, request a new code")
	ErrIncorrectOTP        = errors.New("incorrect otp, try again")
	ErrPasswordBreached    = errors.New("oops the input password was found in a breach, try another password")
)

type infrastructure interface {
	CreateUser(ctx context.Context, input *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CheckPasswordIsBreached(ctx context.Context, password string) (bool, error)
	GetRefreshTokenByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error)
	SaveRefreshToken(ctx context.Context, input domain.RefreshToken) error
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	SendEmailVerification(ctx context.Context, recipient, otp string) (bool, error)
	GenerateOTP(ctx context.Context, userID int64, purpose dto.OTPPurpose) (string, error)
	GetOTPByCodeAndUserID(ctx context.Context, code string, userID int64, purpose dto.OTPPurpose) (*domain.OTP, error)
	MarkOTPAsUsed(ctx context.Context, code string) error
	GetUserByPublicID(ctx context.Context, publicID string) (*domain.User, error)
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

	if u.checkPasswordIsBreached(ctx, input.Password) {
		return nil, ErrPasswordBreached
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

	go func() {
		ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
		defer cancelFunc()

		u.sendOTPMail(ctx, user.ID, user.Email, dto.EmailVerification)
	}()

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

// checkPasswordIsBreached returns true is a password was found in a breach.
func (u *UsecaseImplUser) checkPasswordIsBreached(ctx context.Context, password string) bool {
	ctx, span := telemetry.Trace(ctx, packageName, "checkPasswordIsBreached")
	defer span.End()

	isBreached, err := u.infra.CheckPasswordIsBreached(ctx, password)
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

		return []byte(accessKey), nil
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
	if err != nil || user == nil {
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

	go func() {
		ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
		defer cancelFunc()

		u.sendOTPMail(ctx, user.ID, user.Email, dto.Login)
	}()

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

// sendOTPMail is a utility for sending OTP emails. It returns true on success.
func (u *UsecaseImplUser) sendOTPMail(ctx context.Context, userID int64, recipient string, purpose dto.OTPPurpose) {
	ctx, span := telemetry.Trace(ctx, packageName, "sendMail")
	defer span.End()

	if !purpose.IsValid() {
		slog.ErrorContext(ctx, "invalid type for enum OTPPurpose", "value", purpose.String())

		return
	}

	otp, err := u.infra.GenerateOTP(ctx, userID, purpose)
	if err != nil {
		return
	}

	_, _ = u.infra.SendEmailVerification(ctx, recipient, otp)
}

// VerifyOTP returns true and nil on success.
func (u *UsecaseImplUser) VerifyOTP(ctx context.Context, input *dto.VerifyOTPInput) (bool, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "VerifyOTP")
	defer span.End()

	userID, ok := helpers.GetUserIDCtx(ctx)
	if !ok {
		slog.ErrorContext(ctx, "get userID from context failed")

		return false, errInternalError
	}

	user, err := u.infra.GetUserByPublicID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "get user by public ID failed", "err", err)
		telemetry.RecordError(span, err)

		return false, errInternalError
	}

	otp, err := u.infra.GetOTPByCodeAndUserID(ctx, input.Value, user.ID, input.Purpose)
	if err != nil {
		return false, errInternalError
	}

	if otp == nil {
		return false, ErrInvalidOTPCode
	}

	if !otp.Valid {
		return false, ErrInvalidOTPCode
	}

	// Replace this with a trigger than updates the Valid column.
	if time.Now().After(otp.ExpiresAt) {
		return false, ErrExpiredOTPCode
	}

	if otp.Code != input.Value {
		return false, ErrIncorrectOTP
	}

	go func() {
		ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(1*time.Minute))
		defer cancelFunc()

		err = u.infra.MarkOTPAsUsed(ctx, otp.Code)
		if err != nil {
			slog.ErrorContext(ctx, "revoke otp failed", "err", err)
		}
	}()

	return true, nil
}
