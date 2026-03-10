// Package usecase ...
package usecase

import (
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
)

type Usecase struct {
	*healthcheck.UsecaseImpl
}

func New(health *healthcheck.UsecaseImpl) *Usecase {
	return &Usecase{
		health,
	}
}
