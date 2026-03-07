// Package usecase ...
package usecase

import (
	"github.com/cmmasaba/prototypes/pkg/usecase/healthcheck"
)

type Usecase struct {
	healthcheck *healthcheck.UsecaseImpl
}

func New(health *healthcheck.UsecaseImpl) (*Usecase, error) {
	return &Usecase{
		healthcheck: health,
	}, nil
}
