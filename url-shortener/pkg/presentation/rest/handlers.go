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
	user_uc "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/user"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/presentation/rest"
)

var ErrInvalidRequestBody = errors.New("invalid request body")

type usecases interface {
	CheckDBConnection(ctx context.Context) bool
	CreateUserEmailPassword(ctx context.Context, input *dto.EmailPasswordUserInput) (*dto.User, error)
	ValidatePasswordStrength(ctx context.Context, input *dto.ValidatePasswordInput) bool
	CheckPasswordIsBreached(ctx context.Context, input *dto.ValidatePasswordInput) bool
	Login(ctx context.Context, input *dto.LoginInput) (*dto.LoginResponse, error)
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

	user, err := h.uc.CreateUserEmailPassword(ctx, input)
	if err != nil {
		telemetry.RecordError(span, err)

		switch {
		case errors.Is(err, user_uc.ErrUserWithEmailExists):
			http.Error(w, err.Error(), http.StatusConflict)
		case errors.Is(err, user_uc.ErrIncompleteOAuth):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(user)
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

	pwdStrength := h.uc.ValidatePasswordStrength(ctx, input)
	resp := map[string]bool{
		"passwordStrength": pwdStrength,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		telemetry.RecordError(span, err)
	}
}

// CheckPasswordIsBreached checks is a passowrd was found in a breach.
func (h *Handlers) CheckPasswordIsBreached(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "CheckPasswordIsBreached")
	defer span.End()

	input, ok := decodeAndValidate[dto.ValidatePasswordInput](ctx, r, w)
	if !ok {
		return
	}

	ok = h.uc.CheckPasswordIsBreached(ctx, input)
	resp := map[string]bool{
		"isBreached": ok,
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

	token, err := h.uc.Login(ctx, input)
	if err != nil {
		slog.Error("login failed", "err", err)
		telemetry.RecordError(span, err)

		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		telemetry.RecordError(span, err)
	}
}
