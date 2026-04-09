// Package cache implements caching functionality.
package cache

import (
	"context"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/redis/go-redis/v9"
)

const packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/cache"

type cache interface {
	Get(context.Context, string) *redis.StringCmd
	Set(context.Context, string, any, time.Duration) *redis.StatusCmd
	PoolStats() *redis.PoolStats
}

type Impl struct {
	client cache
}

type PoolStats struct {
	Hits             uint32
	Misses           uint32
	Timeouts         uint32
	TotalConnections uint32
	IdleConnections  uint32
	StaleConnections uint32
}

// New returns a new instance of *[Impl]
func New(client cache) *Impl {
	return &Impl{
		client: client,
	}
}

// Get returns a value by the key from the cache.
func (c *Impl) Get(ctx context.Context, key string) ([]byte, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "Get")
	defer span.End()

	return c.client.Get(ctx, key).Bytes()
}

// Set stores a value by the key in the cache
func (c *Impl) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	ctx, span := telemetry.Trace(ctx, packageName, "Set")
	defer span.End()

	return c.client.Set(ctx, key, value, expiration).Err()
}

// PoolStats returns the underlying client connection pool stats.
func (c *Impl) PoolStats(ctx context.Context) PoolStats {
	_, span := telemetry.Trace(ctx, packageName, "PoolStats")
	defer span.End()

	stats := c.client.PoolStats()

	return PoolStats{
		Hits:             stats.Hits,
		Misses:           stats.Misses,
		Timeouts:         stats.Timeouts,
		TotalConnections: stats.TotalConns,
		IdleConnections:  stats.IdleConns,
		StaleConnections: stats.StaleConns,
	}
}
