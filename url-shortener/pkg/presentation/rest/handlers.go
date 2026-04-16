// Package rest implements handlers for REST API.
package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/presentation/rest"
)

var ErrInvalidRequestBody = errors.New("invalid request body")

const authModeHeader = "X-Auth-Mode"

// useCookieAuth returns true if the client wants cookie-based authentication.
// Cookie mode is used when the client sends the "X-Auth-Mode: cookie" header.
func useCookieAuth(r *http.Request) bool {
	return r.Header.Get(authModeHeader) == "cookie"
}

type usecases interface {
	HealthCheck(ctx context.Context) map[string]bool
	CreateUserEmailPassword(ctx context.Context, input *dto.EmailPasswordUserInput) (*dto.OTPRequiredResponse, error)
	ValidatePasswordStrength(ctx context.Context, input *dto.ValidatePasswordInput) bool
	Login(ctx context.Context, input *dto.LoginInput) (*dto.OTPRequiredResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*dto.RefreshAccessTokenResponse, error)
	VerifyOTP(ctx context.Context, input *dto.VerifyOTPInput) (*dto.AuthResponse, error)
	OAuthFlowCallback(ctx context.Context, origin, code string) (*dto.AuthResponse, error)
	InitOAuthFlow(ctx context.Context, provider dto.OAuthProvider) (string, error)
	Logout(ctx context.Context) error
	RequestNewOTP(ctx context.Context, publicUserID, recipient string, purpose dto.OTPPurpose) error
}

type Handlers struct {
	usecases usecases
	session  *scs.SessionManager
}

func New(usecases usecases, session *scs.SessionManager) *Handlers {
	return &Handlers{
		usecases: usecases,
		session:  session,
	}
}

// decodeAndValidate returns the decoded output and bool.
func decodeAndValidate[T any](ctx context.Context, r *http.Request, w http.ResponseWriter) (*T, bool) {
	ctx, span := telemetry.Trace(ctx, packageName, "decodeAndValidate")
	defer span.End()

	var input T

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.ErrorContext(ctx, "decode request body failed", "err", err)
		telemetry.RecordError(span, err)
		http.Error(w, ErrInvalidRequestBody.Error(), http.StatusBadRequest)

		return nil, false
	}

	if err := helpers.Validate(input); err != nil {
		slog.ErrorContext(ctx, "validate input failed", "err", err)
		telemetry.RecordError(span, err)
		http.Error(w, err.Error(), http.StatusBadRequest)

		return nil, false
	}

	return &input, true
}

// HealthCheck returns the status of backing infrastructure services.
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "HealthCheck")
	defer span.End()

	// do checks for all infra services i.e cache, database, and other critical components.
	resp := h.usecases.HealthCheck(ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		telemetry.RecordError(span, err)
	}
}

// CreateUserEmailPassword persists a new user and initiates OTP verification.
func (h *Handlers) CreateUserEmailPassword(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "CreateUserEmailPassword")
	defer span.End()

	input, ok := decodeAndValidate[dto.EmailPasswordUserInput](ctx, r, w)
	if !ok {
		return
	}

	resp, err := h.usecases.CreateUserEmailPassword(ctx, input)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserWithEmailExists):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, auth.ErrIncompleteOAuth):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, auth.ErrInvalidEmailSyntax):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, auth.ErrPasswordBreached):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	h.session.Put(ctx, "user_id", resp.User.PublicID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}

// ValidatePassword checks if a password entropy is above the threshold.
func (h *Handlers) ValidatePassword(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "ValidatePassword")
	defer span.End()

	input, ok := decodeAndValidate[dto.ValidatePasswordInput](ctx, r, w)
	if !ok {
		return
	}

	ok = h.usecases.ValidatePasswordStrength(ctx, input)
	resp := map[string]bool{
		"passwordStrength": ok,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		telemetry.RecordError(span, err)
	}
}

// Login authenticates credentials and initiates OTP verification.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "Login")
	defer span.End()

	input, ok := decodeAndValidate[dto.LoginInput](ctx, r, w)
	if !ok {
		return
	}

	resp, err := h.usecases.Login(ctx, input)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	h.session.Put(ctx, "user_id", resp.User.PublicID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}

// RefreshAccessToken refreshes access token.
//
// It reads the refresh token from the refresh_token cookie first, falling back to the request body.
func (h *Handlers) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "RefreshAccessToken")
	defer span.End()

	var refreshToken string

	// Try cookie first (browser clients)
	if cookie, err := r.Cookie(refreshTokenCookie); err == nil {
		refreshToken = cookie.Value
	}

	// Fall back to request body (API clients)
	if refreshToken == "" {
		input, ok := decodeAndValidate[dto.RefreshAccessTokenInput](ctx, r, w)
		if !ok {
			return
		}

		refreshToken = input.Token
	}

	resp, err := h.usecases.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredToken) || errors.Is(err, auth.ErrInvalidToken) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	if useCookieAuth(r) {
		SetAuthCookies(w, resp.AccessToken, resp.RefreshToken)

		resp.AccessToken = ""
		resp.RefreshToken = ""
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}

// VerifyOTP validates the OTP using the session and returns JWT tokens on success.
func (h *Handlers) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "VerifyOTP")
	defer span.End()

	userID := h.session.GetString(ctx, "user_id")
	if userID == "" {
		http.Error(w, "session expired or invalid", http.StatusUnauthorized)

		return
	}

	ctx = helpers.SetUserIDCtx(ctx, userID)

	input, ok := decodeAndValidate[dto.VerifyOTPInput](ctx, r, w)
	if !ok {
		return
	}

	resp, err := h.usecases.VerifyOTP(ctx, input)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredOTPCode) || errors.Is(err, auth.ErrIncorrectOTP) ||
			errors.Is(err, auth.ErrInvalidOTPCode) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	err = h.session.Destroy(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "destroy session failed", "err", err)
	}

	if useCookieAuth(r) {
		SetAuthCookies(w, resp.AccessToken, resp.RefreshToken)

		resp.AccessToken = ""
		resp.RefreshToken = ""
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}

// InitGoogleOAuth initiates the Google OAuth flow and redirects the user to the consent page.
func (h *Handlers) InitGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "InitGoogleOAuth")
	defer span.End()

	redirectURL, err := h.usecases.InitOAuthFlow(ctx, dto.OAuthProviderGoogle)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// GoogleOAuthCallback handles the OAuth callback from Google.
func (h *Handlers) GoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "GoogleOAuthCallback")
	defer span.End()

	resp, err := h.usecases.OAuthFlowCallback(ctx, r.URL.String(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	if useCookieAuth(r) {
		SetAuthCookies(w, resp.AccessToken, resp.RefreshToken)

		resp.AccessToken = ""
		resp.RefreshToken = ""
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}

// InitGithubOAuth initiates the Google OAuth flow and redirects the user to the consent page.
func (h *Handlers) InitGithubOAuth(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "InitGithubOAuth")
	defer span.End()

	redirectURL, err := h.usecases.InitOAuthFlow(ctx, dto.OAuthProviderGithub)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// GithubOAuthCallback handles the OAuth callback from GitHub.
func (h *Handlers) GithubOAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "GithubOAuthCallback")
	defer span.End()

	resp, err := h.usecases.OAuthFlowCallback(ctx, r.URL.String(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	if useCookieAuth(r) {
		SetAuthCookies(w, resp.AccessToken, resp.RefreshToken)

		resp.AccessToken = ""
		resp.RefreshToken = ""
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}

// Logout revokes all refresh tokens for the user and clears auth cookies.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "Logout")
	defer span.End()

	if err := h.usecases.Logout(ctx); err != nil {
		slog.ErrorContext(ctx, "logout failed", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	ClearAuthCookies(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"}); err != nil {
		telemetry.RecordError(span, err)
	}
}

func (h *Handlers) RequestNewOTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "RequestNewOTP")
	defer span.End()

	input, ok := decodeAndValidate[dto.RequestOTPInput](ctx, r, w)
	if !ok {
		return
	}

	if err := h.usecases.RequestNewOTP(ctx, input.UserPublicID, input.Recipient, input.Purpose); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
