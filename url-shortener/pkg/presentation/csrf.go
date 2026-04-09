package presentation

import (
	"crypto/subtle"
	"net/http"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
)

// CSRFMiddleware implements the double-submit cookie pattern.
// Safe methods (GET, HEAD, OPTIONS) pass through.
// State-changing methods require the X-CSRF-Token header to match the csrf_token cookie.
func CSRFMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)

				return
			}

			// If the request uses a Bearer token, skip CSRF check.
			if r.Header.Get("Authorization") != "" {
				next.ServeHTTP(w, r)

				return
			}

			cookie, err := r.Cookie(csrfCookieName)
			if err != nil {
				http.Error(w, "CSRF token missing", http.StatusForbidden)

				return
			}

			headerToken := r.Header.Get(csrfHeaderName)
			if headerToken == "" {
				http.Error(w, "CSRF header missing", http.StatusForbidden)

				return
			}

			if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) != 1 {
				http.Error(w, "CSRF token mismatch", http.StatusForbidden)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
