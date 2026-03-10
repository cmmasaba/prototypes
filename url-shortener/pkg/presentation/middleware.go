package presentation

import (
	"log/slog"
	"net/http"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/go-chi/chi/v5/middleware"
)

// LoggingMiddleware facilitates logging across requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := slog.With(
			"request_id", middleware.GetReqID(ctx),
			"method", r.Method,
			"path", r.URL.Path,
		)
		ctx = helpers.SetContextLogger(ctx, logger)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
