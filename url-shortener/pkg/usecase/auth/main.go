// Package auth provides features for auth and user management.
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/tasks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

const (
	packageName           = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
	bcryptCount           = 14
	minEntropy            = 60
	lowercaseCharPoolSize = 26
	uppercaseCharPoolSize = 26
	specialCharPoolSize   = 32
	digitsCharPoolSize    = 10
)

var (
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

	oauthCacheKey = "oauth:%s"
)

type repo interface {
	CreateUser(ctx context.Context, input *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByOAuthProviderAndID(ctx context.Context, oauthProvider, oauthProviderID string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
	GetUserByPublicID(ctx context.Context, publicID string) (*domain.User, error)

	GetRefreshTokenByTokenHash(ctx context.Context, token string) (*domain.RefreshToken, error)
	SaveRefreshToken(ctx context.Context, input domain.RefreshToken) error
	RevokeRefreshToken(ctx context.Context, token string) error
	RevokeAllRefreshTokensForUser(ctx context.Context, userID int64) error
}

type hibp interface {
	CheckPasswordIsBreached(ctx context.Context, password string) (bool, error)
}

type otp interface {
	GenerateOTP(ctx context.Context, userID string, purpose dto.OTPPurpose) (string, error)
	GetOTPByCodeAndUserID(ctx context.Context, code, userPublicID string, purpose dto.OTPPurpose) (*domain.OTP, error)
	RevokeAllOTPsForUser(ctx context.Context, user string, purpose dto.OTPPurpose) error
}

type backgroundTasks interface {
	NewEmailDeliveryTask(ctx context.Context, input tasks.EmailDeliveryPayload, priority tasks.Priority) error
}

type cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
}

type UsecaseImpl struct {
	cache cache
	repo  repo
	hibp  hibp
	otp   otp
	tasks backgroundTasks
}

type oauthState struct {
	Verifier string
	Provider dto.OAuthProvider
}

type googleOAuthUserInfo struct {
	Email string `json:"email"`
	ID    string `json:"id"`
}

type githubOAuthUserInfo struct {
	Email string `json:"email"`
	ID    int64  `json:"id"`
}

func New(repo repo, hibp hibp, otp otp, tasks backgroundTasks, cache cache) *UsecaseImpl {
	return &UsecaseImpl{
		repo:  repo,
		hibp:  hibp,
		otp:   otp,
		tasks: tasks,
		cache: cache,
	}
}

// CreateUserEmailPassword creates a new user and sends an OTP for email verification.
func (u *UsecaseImpl) CreateUserEmailPassword(
	ctx context.Context,
	input *dto.EmailPasswordUserInput,
) (*dto.OTPRequiredResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CreateUserEmailPassword")
	defer span.End()

	e, err := mail.ParseAddress(input.Email)
	if err != nil || e.Address != input.Email {
		slog.ErrorContext(ctx, "validation for email address failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, ErrInvalidEmailSyntax
	}

	existingUser, err := u.repo.GetUserByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		slog.ErrorContext(ctx, "error checking for duplicate email", "err", err)

		return nil, err
	}

	if existingUser != nil {
		err = u.sendMail(ctx, existingUser.PublicID, existingUser.Email, dto.TypeSecurityAlertEmail)
		if err != nil {
			slog.ErrorContext(ctx, "send security alert email failed", "err", err)
		}

		return &dto.OTPRequiredResponse{
			User: dto.User{
				Email: input.Email,
			},
			Status: "otp_required",
		}, nil
	}

	// checkPasswordIsBreached fails open: if HIBP is unreachable, we allow
	// the password. This prioritizes availability over blocking registrations.
	if u.checkPasswordIsBreached(ctx, input.Password) {
		return nil, ErrPasswordBreached
	}

	hashString, err := hashPassword(ctx, input.Password)
	if err != nil {
		return nil, err
	}

	user, err := u.repo.CreateUser(ctx, &domain.User{
		Email:        input.Email,
		PasswordHash: &hashString,
	})
	if err != nil {
		slog.ErrorContext(ctx, "error creating user email/password flow", "err", err)

		return nil, err
	}

	err = u.sendMail(ctx, user.PublicID, user.Email, dto.TypeVerificationEmail)
	if err != nil {
		slog.ErrorContext(ctx, "send email verification otp failed", "err", err)

		return nil, err
	}

	return &dto.OTPRequiredResponse{
		User: dto.User{
			PublicID:  user.PublicID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Status: "otp_required",
	}, nil
}

// ValidatePasswordStrength returns true if the password entropy is >= than minEntropy.
func (u *UsecaseImpl) ValidatePasswordStrength(ctx context.Context, input *dto.ValidatePasswordInput) bool {
	_, span := telemetry.Trace(ctx, packageName, "ValidatePasswordStrength")
	defer span.End()

	base := getBase(input.Password)

	// get number of characters used. len(password) would incorrectly return byte count.
	length := utf8.RuneCountInString(input.Password)

	// Entropy = log2(base^length)
	entropy := float64(length) * math.Log2(float64(base))

	return entropy >= minEntropy
}

// ValidateJWTToken returns the [jwt.MapClaims] and nil error successful after token validation and verification.
func (u *UsecaseImpl) ValidateJWTToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "ValidateJWTToken")
	defer span.End()

	accessKey := helpers.MustGetEnvVar("JWT_ACCESS_SIGNING_KEY")

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}

		return []byte(accessKey), nil
	}, jwt.WithIssuer("url-shortener"), jwt.WithAudience("url-shortener-api"))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
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

// Login verifies credentials and sends an OTP for verification.
func (u *UsecaseImpl) Login(ctx context.Context, input *dto.LoginInput) (*dto.OTPRequiredResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "Login")
	defer span.End()

	user, err := u.repo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		slog.ErrorContext(ctx, "get user by email failed", "err", err)

		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash == nil {
		slog.ErrorContext(ctx, "user does not have login password", "user_id", user.PublicID)

		return nil, ErrInvalidCredentials
	}

	err = verifyPassword(ctx, *user.PasswordHash, input.Password)
	if err != nil {
		slog.ErrorContext(ctx, "password verification failed")
		telemetry.RecordError(span, err)

		return nil, ErrInvalidCredentials
	}

	err = u.sendMail(ctx, user.PublicID, user.Email, dto.TypeLoginEmail)
	if err != nil {
		slog.ErrorContext(ctx, "send login email otp failed", "err", err)

		return nil, err
	}

	return &dto.OTPRequiredResponse{
		User: dto.User{
			PublicID:  user.PublicID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		Status: "otp_required",
	}, nil
}

// Logout returns nil on success.
func (u *UsecaseImpl) Logout(ctx context.Context) error {
	ctx, span := telemetry.Trace(ctx, packageName, "Logout")
	defer span.End()

	userID, ok := helpers.GetUserIDCtx(ctx)
	if !ok {
		slog.ErrorContext(ctx, "get user from context failed")

		return errInternalError
	}

	user, err := u.repo.GetUserByPublicID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "get user by public id failed", "err", err)

		return err
	}

	err = u.repo.RevokeAllRefreshTokensForUser(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "revoke all refresh tokens for user failed", "err", err)

		return err
	}

	return nil
}

// RefreshAccessToken returns a new access token.
func (u *UsecaseImpl) RefreshAccessToken(
	ctx context.Context,
	refreshToken string,
) (*dto.RefreshAccessTokenResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "RefreshAccessToken")
	defer span.End()

	accessTokenTTL, err := strconv.Atoi(helpers.MustGetEnvVar("ACCESS_TOKEN_TTL"))
	if err != nil {
		slog.ErrorContext(ctx, "convert refreshTokenTTL to int failed", "err", err)

		return nil, errInternalError
	}

	inputHash := helpers.HashSecret(refreshToken)

	token, err := u.repo.GetRefreshTokenByTokenHash(ctx, inputHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if token.Revoked {
		return nil, ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	user, err := u.repo.GetUserByID(ctx, token.UserID)
	if err != nil {
		slog.ErrorContext(ctx, "get user by id failed", "err", err)
		return nil, errInternalError
	}

	err = u.repo.RevokeRefreshToken(ctx, inputHash)
	if err != nil {
		slog.ErrorContext(ctx, "revoke refresh token failed", "err", err)

		return nil, errInternalError
	}

	accessToken, newRefreshToken, err := u.generateAuthTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshAccessTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(accessTokenTTL),
	}, nil
}

// VerifyOTP validates the OTP and returns JWT auth tokens on success.
func (u *UsecaseImpl) VerifyOTP(ctx context.Context, input *dto.VerifyOTPInput) (*dto.AuthResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "VerifyOTP")
	defer span.End()

	userID, ok := helpers.GetUserIDCtx(ctx)
	if !ok {
		slog.ErrorContext(ctx, "get userID from context failed")

		return nil, errInternalError
	}

	user, err := u.repo.GetUserByPublicID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "get user by public ID failed", "err", err)

		return nil, errInternalError
	}

	inputHash := helpers.HashSecret(input.Value)

	otp, err := u.otp.GetOTPByCodeAndUserID(ctx, inputHash, userID, input.Purpose)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidOTPCode
		}

		return nil, errInternalError
	}

	if otp.Revoked {
		return nil, ErrInvalidOTPCode
	}

	if time.Now().After(otp.ExpiresAt) {
		return nil, ErrExpiredOTPCode
	}

	err = u.otp.RevokeAllOTPsForUser(ctx, user.PublicID, input.Purpose)
	if err != nil {
		slog.ErrorContext(ctx, "revoke otp failed", "err", err, "user_id", user.PublicID)
	}

	accessToken, refreshToken, err := u.generateAuthTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	accessTokenTTL, _ := strconv.Atoi(helpers.MustGetEnvVar("ACCESS_TOKEN_TTL"))

	return &dto.AuthResponse{
		User: dto.User{
			PublicID:  user.PublicID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL),
	}, nil
}

// InitOAuthFlow implements the OAuth flow for Google
func (u *UsecaseImpl) InitOAuthFlow(ctx context.Context, provider dto.OAuthProvider) (string, error) {
	_, span := telemetry.Trace(ctx, packageName, "InitOAuthFlow")
	defer span.End()

	config := buildOAuthConfig(provider)

	// PKCE protection against CSRF attacks
	verification := oauth2.GenerateVerifier()
	redirectURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verification))

	state := oauthState{Verifier: verification}

	var (
		b        bytes.Buffer
		cacheKey string
	)

	switch provider {
	case dto.OAuthProviderGoogle:
		state.Provider = dto.OAuthProviderGoogle
		cacheKey = fmt.Sprintf(oauthCacheKey, redirectURL)
	case dto.OAuthProviderGithub:
		state.Provider = dto.OAuthProviderGithub
		cacheKey = fmt.Sprintf(oauthCacheKey, redirectURL)
	}

	err := json.NewEncoder(&b).Encode(state)
	if err != nil {
		slog.ErrorContext(ctx, "oauth state encoding failed", "err", err)
		telemetry.RecordError(span, err)

		return "", errInternalError
	}

	err = u.cache.Set(ctx, cacheKey, b.Bytes(), 2*time.Hour)
	if err != nil {
		slog.ErrorContext(ctx, "set oauth PKCE code in cache failed", "err", err)
		telemetry.RecordError(span, err)

		return "", err
	}

	return redirectURL, nil
}

// OAuthFlowCallback handles the OAuth callback from Google and creates the user
func (u *UsecaseImpl) OAuthFlowCallback(ctx context.Context, origin, code string) (*dto.AuthResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "OAuthFlowCallback")
	defer span.End()

	cacheKey := fmt.Sprintf(oauthCacheKey, origin)

	val, err := u.cache.Get(ctx, cacheKey)
	if err != nil {
		slog.ErrorContext(ctx, "get oauth state failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, errInternalError
	}

	var state oauthState

	err = unmarshalJSONToStruct(ctx, bytes.NewBuffer(val), &state)
	if err != nil {
		return nil, errInternalError
	}

	cfg := buildOAuthConfig(state.Provider)

	t, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(state.Verifier))
	if err != nil {
		slog.ErrorContext(ctx, "token exhange for provider failed", "err", err, "provider", state.Provider)
		telemetry.RecordError(span, err)

		return nil, errInternalError
	}

	client := cfg.Client(ctx, t)

	var user *domain.User

	switch state.Provider {
	case dto.OAuthProviderGoogle:
		userInfo, err := u.getGoogleOAuthUserInfo(ctx, client)
		if err != nil {
			return nil, errInternalError
		}

		user, err = u.getOrCreateOAuthUser(ctx, userInfo.Email, state.Provider.String(), userInfo.ID)
		if err != nil {
			slog.ErrorContext(ctx, "getOrCreateOAuthUser failed", "err", err)

			return nil, errInternalError
		}
	case dto.OAuthProviderGithub:
		userInfo, err := u.getGithubOAuthUserInfo(ctx, client)
		if err != nil {
			return nil, errInternalError
		}

		userID := strconv.Itoa(int(userInfo.ID))

		user, err = u.getOrCreateOAuthUser(ctx, userInfo.Email, state.Provider.String(), userID)
		if err != nil {
			slog.ErrorContext(ctx, "create user github oauth flow failed", "err", err)

			return nil, errInternalError
		}
	}

	accessToken, refreshToken, err := u.generateAuthTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	accessTokenTTL, _ := strconv.Atoi(helpers.MustGetEnvVar("ACCESS_TOKEN_TTL"))

	return &dto.AuthResponse{
		User: dto.User{
			PublicID:  user.PublicID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL),
	}, nil
}

func (u *UsecaseImpl) getOrCreateOAuthUser(
	ctx context.Context,
	email, oauthProvider, oauthProviderID string,
) (*domain.User, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "getOrCreateOAuthUser")
	defer span.End()

	user, err := u.repo.GetUserByOAuthProviderAndID(ctx, oauthProvider, oauthProviderID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		slog.ErrorContext(ctx, "get user by oauth provider and oauthid failed", "err", err)

		return nil, err
	}

	if user != nil {
		slog.InfoContext(ctx, "get user by oauth provider and oauthid  found")

		return user, nil
	}

	user, err = u.repo.CreateUser(ctx, &domain.User{
		Email:           email,
		OauthProvider:   &oauthProvider,
		OauthProviderID: &oauthProviderID,
	})
	if err != nil {
		slog.ErrorContext(ctx, "create user by oauth provider and provider id failed", "err", err)

		return nil, err
	}

	return user, nil
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

// sendMail is a utility for sending OTP emails. It returns true on success.
func (u *UsecaseImpl) sendMail(
	ctx context.Context,
	publicUserID string,
	recipient string,
	purpose dto.EmailDeliveryType,
) error {
	ctx, span := telemetry.Trace(ctx, packageName, "sendMail")
	defer span.End()

	if !purpose.IsValid() {
		slog.ErrorContext(ctx, "invalid type for email delivery", "value", purpose)

		return fmt.Errorf("invalid type for email delivery: %v", purpose)
	}

	switch purpose {
	case dto.TypeVerificationEmail, dto.TypeLoginEmail, dto.TypePasswordResetEmail:
		var (
			otp string
			err error
		)

		switch purpose {
		case dto.TypeVerificationEmail:
			otp, err = u.otp.GenerateOTP(ctx, publicUserID, dto.EmailVerification)
		case dto.TypeLoginEmail:
			otp, err = u.otp.GenerateOTP(ctx, publicUserID, dto.Login)
		case dto.TypePasswordResetEmail:
			otp, err = u.otp.GenerateOTP(ctx, publicUserID, dto.PasswordReset)
		}

		if err != nil {
			return err
		}

		err = u.tasks.NewEmailDeliveryTask(ctx, tasks.EmailDeliveryPayload{
			EmailType: purpose,
			Recipient: recipient,
			Opts:      map[string]string{"otpCode": otp},
		},
			tasks.Critical)
		if err != nil {
			return err
		}
	case dto.TypeSecurityAlertEmail:
		err := u.tasks.NewEmailDeliveryTask(ctx, tasks.EmailDeliveryPayload{
			EmailType: purpose,
			Recipient: recipient,
		},
			tasks.Default)
		if err != nil {
			return err
		}
	}

	return nil
}

// generateAuthTokens returns the access and refresh tokens and nil error on success.
func (u *UsecaseImpl) generateAuthTokens(ctx context.Context, user *domain.User) (string, string, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "generateAuthTokens")
	defer span.End()

	accessTokenTTL, err := strconv.Atoi(helpers.MustGetEnvVar("ACCESS_TOKEN_TTL"))
	if err != nil {
		slog.ErrorContext(ctx, "convert accessTokenTTL to int failed", "err", err)

		return "", "", errInternalError
	}

	refreshTokenTTL, err := strconv.Atoi(helpers.MustGetEnvVar("REFRESH_TOKEN_TTL"))
	if err != nil {
		slog.ErrorContext(ctx, "convert refreshTokenTTL to int failed", "err", err)

		return "", "", errInternalError
	}

	accessKey := helpers.MustGetEnvVar("JWT_ACCESS_SIGNING_KEY")
	refreshKey := helpers.MustGetEnvVar("JWT_REFRESH_SIGNING_KEY")
	createdAt := time.Now()
	accessTokenClaims := jwt.MapClaims{
		"exp": createdAt.Add(time.Second * time.Duration(accessTokenTTL)).Unix(),
		"iat": createdAt.Unix(),
		"sub": user.PublicID,
		"jti": uuid.NewString(),
		"iss": "url-shortener",
		"aud": "url-shortener-api",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessTokenClaims)

	accessTokenString, err := accessToken.SignedString([]byte(accessKey))
	if err != nil {
		slog.ErrorContext(ctx, "jwt access token signing failed", "err", err)
		telemetry.RecordError(span, err)

		return "", "", err
	}

	refreshTokenClaims := jwt.MapClaims{
		"sub": user.PublicID,
		"exp": createdAt.Add(time.Second * time.Duration(refreshTokenTTL)).Unix(),
		"jti": uuid.NewString(),
		"iss": "url-shortener",
		"aud": "url-shortener-refresh",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshTokenClaims)

	refreshTokenString, err := refreshToken.SignedString([]byte(refreshKey))
	if err != nil {
		slog.ErrorContext(ctx, "jwt refresh token signing failed", "err", err)
		telemetry.RecordError(span, err)

		return "", "", err
	}

	err = u.repo.SaveRefreshToken(ctx, domain.RefreshToken{
		UserID:    user.ID,
		Token:     helpers.HashSecret(refreshTokenString),
		CreatedAt: createdAt,
		ExpiresAt: createdAt.Add(time.Second * time.Duration(refreshTokenTTL)),
		Revoked:   false,
	})
	if err != nil {
		slog.ErrorContext(ctx, "save refresh token failed", "err", err)

		return "", "", errInternalError
	}

	return accessTokenString, refreshTokenString, nil
}

// hashPassword returns hashed password string and nil if no error occurred.
func hashPassword(ctx context.Context, password string) (string, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "hashPassword")
	defer span.End()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(helpers.HashSecret(password)), bcryptCount)
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

	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(helpers.HashSecret(password)))
}

// checkPasswordIsBreached returns true is a password was found in a breach.
func (u *UsecaseImpl) checkPasswordIsBreached(ctx context.Context, password string) bool {
	ctx, span := telemetry.Trace(ctx, packageName, "checkPasswordIsBreached")
	defer span.End()

	isBreached, err := u.hibp.CheckPasswordIsBreached(ctx, password)
	if err != nil {
		slog.WarnContext(ctx, "checking password breach skipped", "err", err)
		telemetry.RecordError(span, err)

		return isBreached
	}

	return isBreached
}

func buildOAuthConfig(provider dto.OAuthProvider) *oauth2.Config {
	switch provider {
	case dto.OAuthProviderGoogle:
		return &oauth2.Config{
			ClientID:     helpers.MustGetEnvVar("GOOGLE_OAUTH_CLIENT_ID"),
			ClientSecret: helpers.MustGetEnvVar("GOOGLE_OAUTH_CLIENT_SECRET"),
			RedirectURL:  helpers.MustGetEnvVar("GOOGLE_OAUTH_REDIRECT_URL"),
			Scopes:       strings.Split(helpers.MustGetEnvVar("GOOGLE_OAUTH_SCOPES"), ","),
			Endpoint:     google.Endpoint,
		}
	case dto.OAuthProviderGithub:
		return &oauth2.Config{
			ClientID:     helpers.MustGetEnvVar("GITHUB_OAUTH_CLIENT_ID"),
			ClientSecret: helpers.MustGetEnvVar("GITHUB_OAUTH_CLIENT_SECRET"),
			RedirectURL:  helpers.MustGetEnvVar("GITHUB_OAUTH_REDIRECT_URL"),
			Scopes:       strings.Split(helpers.MustGetEnvVar("GITHUB_OAUTH_SCOPES"), ","),
			Endpoint:     github.Endpoint,
		}
	default:
		return nil
	}
}

func unmarshalJSONToStruct(ctx context.Context, src io.Reader, dest any) error {
	ctx, span := telemetry.Trace(ctx, packageName, "unmarshalJSONToStruct")
	defer span.End()

	err := json.NewDecoder(src).Decode(dest)
	if err != nil {
		slog.ErrorContext(ctx, "decode json to struct failed", "err", err)
		telemetry.RecordError(span, err)

		return errInternalError
	}

	return nil
}

func (u *UsecaseImpl) getGoogleOAuthUserInfo(ctx context.Context, client *http.Client) (*googleOAuthUserInfo, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "getGoogleOAuthUserInfo")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, helpers.MustGetEnvVar("GOOGLE_USER_INFO_API"), nil)
	if err != nil {
		slog.ErrorContext(ctx, "build google profile api request failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, errInternalError
	}

	resp, err := client.Do(req) // nolint: gosec
	if err != nil {
		slog.ErrorContext(ctx, "http request to get user info failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, errInternalError
	}

	defer resp.Body.Close()

	var userInfo googleOAuthUserInfo

	err = unmarshalJSONToStruct(ctx, resp.Body, &userInfo)
	if err != nil {
		return nil, errInternalError
	}

	return &userInfo, nil
}

func (u *UsecaseImpl) getGithubOAuthUserInfo(ctx context.Context, client *http.Client) (*githubOAuthUserInfo, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "getGithubOAuthUserInfo")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, helpers.MustGetEnvVar("GITHUB_USER_INFO_API"), nil)
	if err != nil {
		slog.ErrorContext(ctx, "build github profile api request failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, errInternalError
	}

	resp, err := client.Do(req) // nolint: gosec
	if err != nil {
		slog.ErrorContext(ctx, "http request to get user info failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, errInternalError
	}

	defer resp.Body.Close()

	var userInfo githubOAuthUserInfo

	err = unmarshalJSONToStruct(ctx, resp.Body, &userInfo)
	if err != nil {
		return nil, errInternalError
	}

	return &userInfo, nil
}
