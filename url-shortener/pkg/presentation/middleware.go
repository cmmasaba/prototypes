package presentation

import (
	"net/http"
	"strings"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/presentation"
)

// AuthMiddleware extracts Authorization header from the request and validates it
func AuthMiddleware(user *auth.UsecaseImpl) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := telemetry.Trace(r.Context(), packageName, "AuthMiddleware")
			defer span.End()

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)

				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization format", http.StatusUnauthorized)

				return
			}

			tokenString := parts[1]

			claims, err := user.ValidateJWTToken(ctx, tokenString)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)

				return
			}

			userID, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)

				return
			}

			ctx = helpers.SetUserIDCtx(ctx, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
