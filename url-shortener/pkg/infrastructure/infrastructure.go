// Package infrastructure facilitates interaction with repository and external services
package infrastructure

import (
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/hibp"
)

type Infrastructure struct {
	*repository.Repository
	*hibp.HIBP
}

func New(database *repository.Repository, hibp *hibp.HIBP) (*Infrastructure, error) {
	return &Infrastructure{
		database,
		hibp,
	}, nil
}
