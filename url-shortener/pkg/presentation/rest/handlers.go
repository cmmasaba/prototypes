// Package rest implements handlers for REST API.
package rest

import (
	"context"
	"net/http"

	"github.com/cmmasaba/prototypes/telemetry"
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("system up"))
	} else {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("system down"))
	}
}
