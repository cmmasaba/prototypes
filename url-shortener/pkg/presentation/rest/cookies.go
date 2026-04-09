package rest

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
)

const (
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
	csrfTokenCookie    = "csrf_token"
)

// SetAuthCookies sets HTTP-only secure cookies for access and refresh tokens,
// and a non-HttpOnly CSRF token cookie for the double-submit pattern.
func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken string) {
	accessTokenTTL := helpers.MustGetEnvVar("ACCESS_TOKEN_TTL")
	refreshTokenTTL := helpers.MustGetEnvVar("REFRESH_TOKEN_TTL")

	accessTokenMaxAge, err := strconv.Atoi(accessTokenTTL)
	if err != nil {
		slog.Error("convert access token ttl to int failed", "err", err)

		return
	}

	refreshTokenMaxAge, err := strconv.Atoi(refreshTokenTTL)
	if err != nil {
		slog.Error("convert refresh token ttl to int failed", "err", err)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     accessTokenCookie,
		Value:    accessToken,
		Path:     "/api",
		MaxAge:   accessTokenMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    refreshToken,
		Path:     "/api/auth/refresh",
		MaxAge:   refreshTokenMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	csrfBytes := make([]byte, 32)
	if _, err := rand.Read(csrfBytes); err != nil {
		slog.Error("failed to generate CSRF token", "err", err)

		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     csrfTokenCookie,
		Value:    hex.EncodeToString(csrfBytes),
		Path:     "/",
		MaxAge:   accessTokenMaxAge,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAuthCookies removes auth and CSRF cookies by setting MaxAge to -1.
func ClearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessTokenCookie,
		Value:    "",
		Path:     "/api",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    "",
		Path:     "/api/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     csrfTokenCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}
