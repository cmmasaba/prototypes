// Package rest implements handlers for REST API.
package rest

import (
	"context"
	"encoding/json"
	"errors"
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
	CheckDBConnection(ctx context.Context) error
	CreateUserEmailPassword(ctx context.Context, input dto.EmailPasswordUserInput) (*dto.UserOutput, error)
}

type Handlers struct {
	uc usecases
}

func New(uc usecases) *Handlers {
	return &Handlers{
		uc: uc,
	}
}

// HealthCheck returns the status of backing infrastructure services.
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "HealthCheck")
	defer span.End()

	err := h.uc.CheckDBConnection(ctx)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)

		if _, err := w.Write([]byte(`{"status": "down"}`)); err != nil {
			telemetry.RecordError(span, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(`{"status": "up"}`)); err != nil {
		telemetry.RecordError(span, err)
	}
}

// CreateUserEmailPassword persists a new user to the database.
func (h *Handlers) CreateUserEmailPassword(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "CreateUserEmailPassword")
	defer span.End()

	var input dto.EmailPasswordUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
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
		case errors.Is(err, user_uc.ErrInvalidAuthMethod),
			errors.Is(err, user_uc.ErrNoAuthMethod),
			errors.Is(err, user_uc.ErrIncompleteOAuth):
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
