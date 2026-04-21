// Package shortener implements url shortening functionalities
package shortener

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/shortener"
	alphabet    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

var (
	ErrInvalidURIFormat = errors.New("invalid url string provided")
	ErrURLTooLong       = errors.New("url length exceeds maximum length of 2048 characters")
	errInternalError    = errors.New("internal error")
	ErrURLNotFound      = errors.New("url matching short link not found")
	ErrURLExpired       = errors.New("short link has expired")

	serviceURL = helpers.MustGetEnvVar("SERVICE_BASE_URL")
)

type repo interface {
	GetLinkByCode(ctx context.Context, code string) (*domain.Link, error)
	CreateShortLink(ctx context.Context, input domain.Link) (*domain.Link, error)
}

type cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
}

type UsecaseImpl struct {
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

// New returns an instance of *[UsecaseImpl]
func New(repo repo, cache cache) *UsecaseImpl {
	return &UsecaseImpl{
		repo:  repo,
		cache: cache,
	}
}

func validateURL(rawURL string) error {
	if utf8.RuneCountInString(rawURL) > 2048 {
		return ErrURLTooLong
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Host == "" || u.Scheme == "" {
		return ErrInvalidURIFormat
	}

	return nil
}

// ShortenURL returns *[dto.ShortenURLResponse] and a nil error on success.
func (s *UsecaseImpl) ShortenURL(ctx context.Context, input *dto.ShortenURLInput) (*dto.ShortenURLResponse, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "ShortenURL")
	defer span.End()

	err := validateURL(input.URL)
	if err != nil {
		telemetry.RecordError(span, err)
		slog.ErrorContext(ctx, "url validation failed", "rawURL", input.URL)

		return nil, err
	}

	user, ok := helpers.GetUserIDCtx(ctx)
	if !ok {
		telemetry.RecordError(span, fmt.Errorf("get user from context failed"))
		slog.ErrorContext(ctx, "get user from context failed", "rawURL", input.URL)

		return nil, errInternalError
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

		var data domain.Link

		if user == dto.AnonymousUserID {
			ownershipToken := helpers.HashSecret(code)

			data = domain.Link{
				ShortCode:      code,
				OriginalURL:    input.URL,
				OwnershipToken: &ownershipToken,
			}
		} else {
			data = domain.Link{
				ShortCode:   code,
				UserID:      user,
				OriginalURL: input.URL,
			}

			if input.ExpiresAt.IsZero() {
				data.ExpiresAt = nil
			} else {
				data.ExpiresAt = &input.ExpiresAt
			}
		}

		link, err := s.repo.CreateShortLink(ctx, data)
		if err != nil {
			telemetry.RecordError(span, err)
			slog.ErrorContext(ctx, "create short code failed", "err", err)

			return nil, errInternalError
		}

		resp := &dto.ShortenURLResponse{
			ShortURL: fmt.Sprintf("%s/%s", serviceURL, link.ShortCode),
		}

		if link.OwnershipToken == nil {
			resp.OwnershipToken = ""
		} else {
			resp.OwnershipToken = *link.OwnershipToken
		}

		return resp, nil
	}

	return nil, errInternalError
}

// GetOriginalURL returns the original URL for the given short code.
func (s *UsecaseImpl) GetOriginalURL(ctx context.Context, code string) (string, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetOriginalURL")
	defer span.End()

	cacheKey := fmt.Sprintf("shortcode:%s", code)

	cached, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var link domain.Link

		err := json.NewDecoder(bytes.NewReader(cached)).Decode(&link)
		if err != nil {
			slog.ErrorContext(ctx, "decode cached link data failed", "err", err)
			telemetry.RecordError(span, err)

			return "", errInternalError
		}

		return link.OriginalURL, nil
	}

	link, err := s.repo.GetLinkByCode(ctx, code)
	if err != nil {
		slog.ErrorContext(ctx, "get link by shortcode failed", "err", err)
		telemetry.RecordError(span, err)

		return "", ErrURLNotFound
	}

	if time.Now().After(*link.ExpiresAt) {
		slog.ErrorContext(ctx, "short link expired", "code", code)

		return "", ErrURLExpired
	}

	var b bytes.Buffer

	// Shouldn't fail on error since cache is an optimization.
	err = json.NewEncoder(&b).Encode(link)
	if err != nil {
		slog.ErrorContext(ctx, "encode link data to bytes failed", "err", err)
		telemetry.RecordError(span, err)

		return link.OriginalURL, nil
	}

	err = s.cache.Set(ctx, cacheKey, b, 1*time.Hour)
	if err != nil {
		slog.ErrorContext(ctx, "save link data to cache failed", "err", err)
		telemetry.RecordError(span, err)
	}

	return link.OriginalURL, nil
}
