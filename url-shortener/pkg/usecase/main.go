// Package usecase ...
package usecase

import (
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/user"
)

type Usecase struct {
	*healthcheck.UsecaseImplHealth
	*user.UsecaseImplUser
}

func New(health *healthcheck.UsecaseImplHealth, user *user.UsecaseImplUser) *Usecase {
	return &Usecase{
		health,
		user,
	}
}
