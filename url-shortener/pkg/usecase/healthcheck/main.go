// Package healthcheck...
package healthcheck

import (
	"context"

	"github.com/cmmasaba/prototypes/telemetry"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
)

type usecase interface {
	PingDB(context.Context) error
}

type UsecaseImplHealth struct {
	infra usecase
}

func New(infrastructure usecase) *UsecaseImplHealth {
	return &UsecaseImplHealth{
		infra: infrastructure,
	}
}

// CheckDBConnection returns true after pinging db connection.
func (u *UsecaseImplHealth) CheckDBConnection(ctx context.Context) bool {
	ctx, span := telemetry.Trace(ctx, packageName, "CheckDBConnection")
	defer span.End()

	err := u.infra.PingDB(ctx)
	if err != nil {
		telemetry.RecordError(span, err)

		return false
	}

	return true
}
