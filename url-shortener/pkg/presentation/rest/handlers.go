// Package rest implements handlers for REST API.
package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	user_uc "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/user"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/presentation/rest"
)

var ErrInvalidRequestBody = errors.New("invalid request body")

type usecases interface {
	CheckDBConnection(ctx context.Context) bool
	CreateUserEmailPassword(ctx context.Context, input dto.EmailPasswordUserInput) (*dto.UserOutput, error)
	ValidatePasswordStrength(ctx context.Context, input dto.ValidatePasswordInput) bool
	CheckPasswordIsBreached(ctx context.Context, input dto.ValidatePasswordInput) bool
}

type Handlers struct {
	uc usecases
}

func New(uc usecases) *Handlers {
	return &Handlers{
		uc: uc,
	}
}

func decodeBody(src io.Reader, dest any) error {
	if err := json.NewDecoder(src).Decode(&dest); err != nil {
		return err
	}

	return nil
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

	var input dto.EmailPasswordUserInput

	err := decodeBody(r.Body, input)
	if err != nil {
		telemetry.RecordError(span, err)
		http.Error(w, ErrInvalidRequestBody.Error(), http.StatusBadRequest)

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

	var input dto.ValidatePasswordInput

	err := decodeBody(r.Body, &input)
	if err != nil {
		telemetry.RecordError(span, err)
		http.Error(w, ErrInvalidRequestBody.Error(), http.StatusBadRequest)

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

	var input dto.ValidatePasswordInput

	err := decodeBody(r.Body, &input)
	if err != nil {
		telemetry.RecordError(span, err)
		http.Error(w, ErrInvalidRequestBody.Error(), http.StatusBadRequest)

		return
	}

	ok := h.uc.CheckPasswordIsBreached(ctx, input)
	resp := map[string]bool{
		"breached": ok,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		telemetry.RecordError(span, err)
	}
}
