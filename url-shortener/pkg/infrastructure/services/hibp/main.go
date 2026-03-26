// Package hibp verifies passwords against known breaches using the Have I Been Pwned API.
package hibp

import (
	"context"
	"crypto/sha1" // nolint: gosec
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/pwnedapi"
)

// HIBP encapsulates the API for checking password breaches.
type HIBP struct {
	client  *http.Client
	baseURL string
}

// New returns an instance of *[HIBP]
func New() (*HIBP, error) {
	baseURL := helpers.MustGetEnvVar("PWNED_API_URL")

	return &HIBP{
		client: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   10 * time.Second,
		},
		baseURL: baseURL,
	}, nil
}

// CheckPasswordIsBreached returns true if the password shows up in a brach database.
func (h *HIBP) CheckPasswordIsBreached(ctx context.Context, password string) (bool, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CheckPasswordIsBreached")
	defer span.End()

	hasher := sha1.New() // nolint: gosec

	hasher.Write([]byte(password))

	pwdHash := hex.EncodeToString(hasher.Sum(nil))
	frange, lrange := pwdHash[0:5], pwdHash[5:40]
	fullURL := h.baseURL + frange

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		telemetry.RecordError(span, err)
		slog.ErrorContext(ctx, "error building request", "err", err)

		return false, err
	}

	resp, err := h.client.Do(req) // nolint: gosec
	if err != nil {
		telemetry.RecordError(span, err)
		slog.ErrorContext(ctx, "error making request", "err", err)

		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.WarnContext(ctx, "pwned API returned non-200 status", "status", resp.StatusCode) // nolint: gosec

		return false, fmt.Errorf("pwned API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		telemetry.RecordError(span, err)
		slog.ErrorContext(ctx, "error decoding response body", "err", err)

		return false, err
	}

	for resp := range strings.SplitSeq(string(body), "\r\n") {
		parts := strings.SplitN(resp, ":", 2)
		if len(parts) != 2 {
			continue
		}

		if parts[0] == lrange {
			return true, nil
		}
	}

	return false, nil
}
