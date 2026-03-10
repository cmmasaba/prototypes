// Package healthcheck...
package healthcheck

import (
	"context"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
)

type Usecase interface {
	PingDB(context.Context) bool
}

type UsecaseImpl struct {
	Usecase
}

func New(infrastructure infrastructure.Infrastructure) *UsecaseImpl {
	return &UsecaseImpl{
		Usecase: infrastructure,
	}
}

func (u *UsecaseImpl) CheckDBConnection(ctx context.Context) bool {
	ctx, span := telemetry.Trace(ctx, packageName, "CheckDBConnection")
	defer span.End()

	return u.PingDB(ctx)
}
