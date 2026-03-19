// Package rest implements handlers for REST API.
package rest

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/presentation/rest"
)

var ErrInvalidRequestBody = errors.New("invalid request body")

type usecases interface {
	CheckDBConnection(ctx context.Context) bool
	CreateUserEmailPassword(ctx context.Context, input *dto.EmailPasswordUserInput) (*dto.AuthResponse, error)
	ValidatePasswordStrength(ctx context.Context, input *dto.ValidatePasswordInput) bool
	Login(ctx context.Context, input *dto.LoginInput) (*dto.AuthResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*dto.RefreshAccessTokenResponse, error)
	VerifyOTP(ctx context.Context, input *dto.VerifyOTPInput) (bool, error)
}

type Handlers struct {
	uc usecases
}

func New(uc usecases) *Handlers {
	return &Handlers{
		uc: uc,
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
	ok := h.uc.CheckDBConnection(ctx)

	resp := map[string]bool{
		"db": ok,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		telemetry.RecordError(span, err)
	}
}

// CreateUserEmailPassword persists a new user to the database.
func (h *Handlers) CreateUserEmailPassword(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "CreateUserEmailPassword")
	defer span.End()

	input, ok := decodeAndValidate[dto.EmailPasswordUserInput](ctx, r, w)
	if !ok {
		return
	}

	resp, err := h.uc.CreateUserEmailPassword(ctx, input)
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

	ok = h.uc.ValidatePasswordStrength(ctx, input)
	resp := map[string]bool{
		"passwordStrength": ok,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		telemetry.RecordError(span, err)
	}
}

// Login authenticates the user by email/password.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "Login")
	defer span.End()

	input, ok := decodeAndValidate[dto.LoginInput](ctx, r, w)
	if !ok {
		return
	}

	resp, err := h.uc.Login(ctx, input)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}

// RefreshAccessToken refreshes access tokens.
func (h *Handlers) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "RefreshAccessToken")
	defer span.End()

	input, ok := decodeAndValidate[dto.RefreshAccessTokenInput](ctx, r, w)
	if !ok {
		return
	}

	resp, err := h.uc.RefreshAccessToken(ctx, input.Token)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredToken) || errors.Is(err, auth.ErrInvalidToken) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}

// VerifyOTP does OTP verification.
func (h *Handlers) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "VerifyOTP")
	defer span.End()

	input, ok := decodeAndValidate[dto.VerifyOTPInput](ctx, r, w)
	if !ok {
		return
	}

	ok, err := h.uc.VerifyOTP(ctx, input)
	if err != nil {
		if errors.Is(err, auth.ErrExpiredOTPCode) || errors.Is(err, auth.ErrIncorrectOTP) || errors.Is(err, auth.ErrInvalidOTPCode) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	resp := map[string]bool{
		"verified": ok,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}
