// Package shortener implements url shortening functionalities
package shortener

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"unicode/utf8"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/jxskiss/base62"
)

const packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/shortener"

var (
	errInvalidURIFormat = errors.New("invalid url string provided")
	errURLTooLong       = errors.New("url length exceeds maximum length of 2048 characters")
	errInternalError    = errors.New("internal error")
)

type repo interface {
	GetLinkByCode(ctx context.Context, code string) (*domain.Link, error)
	CreateShortLink(ctx context.Context, input domain.Link) (*domain.Link, error)
}

type Shortener struct {
	repo repo
}

func New(repo repo) *Shortener {
	return &Shortener{
		repo: repo,
	}
}

func (s *Shortener) isCodeDuplicate(ctx context.Context, code string) (bool, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "isCodeDuplicate")
	defer span.End()

	existingCode, err := s.repo.GetLinkByCode(ctx, code)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		telemetry.RecordError(span, err)
		slog.ErrorContext(ctx, "check for duplicate short code failed", "err", err)

		return false, errInternalError
	}

	if existingCode != nil {
		slog.WarnContext(ctx, "short code collision", "code", code)

		return true, nil
	}

	return false, nil
}

func (s *Shortener) ShortenURL(ctx context.Context, rawURL string) (*domain.Link, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "EncodeURL")
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

	code := base62.EncodeToString([]byte(rawURL))
	slog.InfoContext(ctx, "generated", "code", code)

	// check for duplicate and retry 5 times before giving up
	for retries := range 5 {
		isDuplicate, err := s.isCodeDuplicate(ctx, code)
		if err != nil {
			return nil, err
		}

		if !isDuplicate {
			break
		}

		if isDuplicate && retries != 4 {
			code = base62.EncodeToString([]byte(rawURL))
			continue
		}

		return nil, errInternalError
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

func (s *Shortener) DecodeURL(ctx context.Context) {
	_, span := telemetry.Trace(ctx, packageName, "DecodeURL")
	defer span.End()
}
