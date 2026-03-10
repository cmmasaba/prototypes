// Package helpers contains common utilities
package helpers

import (
	"context"
	"log/slog"
)

type ctxKey string

const loggerKey ctxKey = "Logger"

func GetContextLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}

	return slog.Default()
}

func SetContextLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
