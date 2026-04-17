package presentation

import (
	"net/http"
	"strings"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
)

// AuthMiddleware extracts the access token from the Authorization header or the access_token cookie
// and validates it. Bearer header takes priority over cookie for API client compatibility.
func AuthMiddleware(auth *auth.UsecaseImpl) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			var tokenString string

			// Try Authorization header first (API clients)
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || parts[0] != "Bearer" {
					http.Error(w, "invalid authorization format", http.StatusUnauthorized)

					return
				}

				tokenString = parts[1]
			}

			// Fall back to access_token cookie (browser clients)
			if tokenString == "" {
				cookie, err := r.Cookie("access_token")
				if err != nil {
					http.Error(w, "authentication required", http.StatusUnauthorized)

					return
				}

				tokenString = cookie.Value
			}

			claims, err := auth.ValidateJWTToken(ctx, tokenString)
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

// OptionalAuthMiddleware attempts to extract an access token from the request and if it fails
// allows users to proceed in anonymous mode.
func OptionalAuthMiddleware(auth *auth.UsecaseImpl) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			var tokenString string

			// Try Authorization header first (API clients)
			if authHeader := r.Header.Get("Authorization"); authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) != 2 || parts[0] != "Bearer" {
					http.Error(w, "invalid authorization format", http.StatusUnauthorized)

					return
				}

				tokenString = parts[1]
			}

			// Fall back to access_token cookie (browser clients)
			if tokenString == "" {
				cookie, err := r.Cookie("access_token")
				if err != nil {
					http.Error(w, "authentication required", http.StatusUnauthorized)

					return
				}

				tokenString = cookie.Value
			}

			if tokenString != "" {
				claims, err := auth.ValidateJWTToken(ctx, tokenString)
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
			} else {
				ctx = helpers.SetUserIDCtx(ctx, dto.AnonymousUserID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
