// Package healthcheck returns the status of infra components.
package healthcheck

import (
	"context"

	"github.com/cmmasaba/prototypes/telemetry"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
)

type repository interface {
	PingDB(context.Context) bool
}

type queue interface {
	PingQueue(ctx context.Context) bool
}

type UsecaseImpl struct {
	repo      repository
	taskqueue queue
}

func New(repo repository, queue queue) *UsecaseImpl {
	return &UsecaseImpl{
		repo:      repo,
		taskqueue: queue,
	}
}

// HealthCheck returns a map of infra service and a boolean indicating if it was reachable.
func (u *UsecaseImpl) HealthCheck(ctx context.Context) map[string]bool {
	ctx, span := telemetry.Trace(ctx, packageName, "HealthCheck")
	defer span.End()

	dbStatus := u.repo.PingDB(ctx)
	queueStatus := u.taskqueue.PingQueue(ctx)

	return map[string]bool{
		"database":   dbStatus,
		"tasksqueue": queueStatus,
	}
}
