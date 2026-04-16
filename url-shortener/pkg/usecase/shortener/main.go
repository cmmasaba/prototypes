// Package shortener implements url shortening functionalities
package shortener

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/shortener"
	alphabet    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	errInvalidURIFormat = errors.New("invalid url string provided")
	errURLTooLong       = errors.New("url length exceeds maximum length of 2048 characters")
	errInternalError    = errors.New("internal error")
)

type repo interface {
	GetLinkByCode(ctx context.Context, code string) (*domain.Link, error)
	CreateShortLink(ctx context.Context, input domain.Link) (*domain.Link, error)
}

type cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
}

type Shortener struct {
	repo  repo
	cache cache
}

func base62Encode(n uint64) string {
	if n == 0 {
		return string(alphabet[0])
	}

	var sb strings.Builder

	for n > 0 {
		sb.WriteByte(alphabet[n%62])
		n /= 62
	}

	// reverse
	result := []byte(sb.String())

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

func New(repo repo, cache cache) *Shortener {
	return &Shortener{
		repo:  repo,
		cache: cache,
	}
}

func (s *Shortener) ShortenURL(ctx context.Context, rawURL string) (*domain.Link, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "ShortenURL")
	defer span.End()

	charCount := utf8.RuneCountInString(rawURL)

	if charCount > 2048 {
		telemetry.RecordError(span, errURLTooLong)
		slog.ErrorContext(ctx, "url length exceeds maximum length", "length", charCount, "rawURL", rawURL)

		return nil, errURLTooLong
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Host == "" {
		telemetry.RecordError(span, errInvalidURIFormat)
		slog.ErrorContext(ctx, "invalid url format", "rawURL", rawURL)

		return nil, errInvalidURIFormat
	}

	for range 5 {
		b := make([]byte, 8)

		_, _ = rand.Read(b)

		code := base62Encode(binary.BigEndian.Uint64(b))[:7]

		existingCode, err := s.repo.GetLinkByCode(ctx, code)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			telemetry.RecordError(span, err)
			slog.ErrorContext(ctx, "check for duplicate short code failed", "err", err)

			continue
		}

		if existingCode != nil {
			slog.InfoContext(ctx, "short code collision detected", "code", code)

			continue
		}

		link, err := s.repo.CreateShortLink(ctx, domain.Link{
			ShortCode:   code,
			OriginalURL: rawURL,
		})
		if err != nil {
			telemetry.RecordError(span, err)
			slog.ErrorContext(ctx, "save short code failed", "err", err)

			return nil, errInternalError
		}

		return link, nil
	}

	return nil, errInternalError
}

func (s *Shortener) DecodeURL(ctx context.Context) {
	_, span := telemetry.Trace(ctx, packageName, "DecodeURL")
	defer span.End()
}
