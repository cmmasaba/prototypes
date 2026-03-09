// Package infrastructure facilitates interaction with repository and external services
package infrastructure

import (
	"context"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
)

type Infrastructure interface {
	PingDB(context.Context) bool
}

type InfraImpl struct {
	r *repository.Repository
}

func New() (Infrastructure, error) {
	r, err := repository.New()
	if err != nil {
		return nil, err
	}

	return &InfraImpl{
		r: r,
	}, nil
}

func (i *InfraImpl) PingDB(ctx context.Context) bool {
	err := i.r.Ping(ctx)

	return err == nil
}
