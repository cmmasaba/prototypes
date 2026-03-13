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

type UsecaseImpl struct {
	infra usecase
}

func New(infrastructure usecase) *UsecaseImpl {
	return &UsecaseImpl{
		infra: infrastructure,
	}
}

func (u *UsecaseImpl) CheckDBConnection(ctx context.Context) error {
	ctx, span := telemetry.Trace(ctx, packageName, "CheckDBConnection")
	defer span.End()

	return u.infra.PingDB(ctx)
}
