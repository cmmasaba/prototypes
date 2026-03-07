// Package healthcheck...
package healthcheck

import (
	"context"

	"github.com/cmmasaba/prototypes/pkg/infrastructure"
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
	return u.PingDB(ctx)
}
