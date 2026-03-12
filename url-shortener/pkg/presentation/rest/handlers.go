// Package rest implements handlers for REST API.
package rest

import (
	"context"
	"net/http"

	"github.com/cmmasaba/prototypes/telemetry"
	"go.opentelemetry.io/otel/codes"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/presentation/rest"
)

type usecases interface {
	CheckDBConnection(ctx context.Context) bool
}

type Handlers struct {
	uc usecases
}

func New(uc usecases) *Handlers {
	return &Handlers{
		uc: uc,
	}
}

// HealthCheck returns the health status of the server and backing infrastructure services.
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, span := telemetry.Trace(r.Context(), packageName, "HealthCheck")
	defer span.End()

	ok := h.uc.CheckDBConnection(ctx)

	if ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte(`{"status": "up"}`)); err != nil {
			span.SetStatus(codes.Error, "an error occured")
			span.RecordError(err)
		}
	} else {
		w.WriteHeader(http.StatusBadGateway)

		if _, err := w.Write([]byte(`{"status": "down"}`)); err != nil {
			span.SetStatus(codes.Error, "an error occured")
			span.RecordError(err)
		}
	}
}
