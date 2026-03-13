// Package infrastructure facilitates interaction with repository and external services
package infrastructure

import (
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
)

type Infrastructure struct {
	*repository.Repository
}

func New(database *repository.Repository) (*Infrastructure, error) {
	return &Infrastructure{
		database,
	}, nil
}
