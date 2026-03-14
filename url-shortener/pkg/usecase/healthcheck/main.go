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

func (u *UsecaseImplHealth) CheckDBConnection(ctx context.Context) error {
	ctx, span := telemetry.Trace(ctx, packageName, "CheckDBConnection")
	defer span.End()

	return u.infra.PingDB(ctx)
}
