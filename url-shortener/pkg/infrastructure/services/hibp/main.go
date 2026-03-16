// Package hibp verifies passwords against known breaches using the Have I Been Pwned API.
package hibp

import (
	"context"
	"crypto/sha1" // nolint: gosec
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/pwnedapi"
)

var baseURL = os.Getenv("PWNED_API_URL")

// HIBP encapsulates the API for checking password breaches.
type HIBP struct {
	client http.Client
}

// New returns an instance of *[PwnedAPI]
func New() (*HIBP, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("pwned api url not set")
	}

	return &HIBP{
		client: http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   10 * time.Second,
		},
	}, nil
}

// CheckPasswordIsBreached returns true if the password shows up in a brach database.
func (p *HIBP) CheckPasswordIsBreached(ctx context.Context, password string) (bool, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "CheckPasswordIsBreached")
	defer span.End()

	hasher := sha1.New() // nolint: gosec

	hasher.Write([]byte(password))

	pwdHash := fmt.Sprintf("%X", hasher.Sum(nil))
	frange, lrange := pwdHash[0:5], pwdHash[5:40]
	fullURL := baseURL + frange

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		telemetry.RecordError(span, err)
		slog.Error("error building request", "err", err)

		return false, err
	}

	resp, err := p.client.Do(req) // nolint: gosec
	if err != nil {
		telemetry.RecordError(span, err)
		slog.Error("error making request", "err", err)

		return false, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn("pwned API returned non-200 status", "status", resp.StatusCode) // nolint: gosec

		return false, fmt.Errorf("pwned API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		telemetry.RecordError(span, err)
		slog.Error("error decoding response body", "err", err)

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
