// Package usecase encapsulates all usecase objects.
package usecase

import (
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
)

type Usecase struct {
	*healthcheck.UsecaseImplHealth
	*auth.UsecaseImplUser
}

func New(health *healthcheck.UsecaseImplHealth, auth *auth.UsecaseImplUser) *Usecase {
	return &Usecase{
		health,
		auth,
	}
}
