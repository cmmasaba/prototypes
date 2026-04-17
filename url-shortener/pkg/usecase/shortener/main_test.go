package shortener

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/shortener/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestShortener_ShortenURL(t *testing.T) {
	var sb strings.Builder
	for range 2500 {
		sb.WriteString("a")
	}

	type args struct {
		input *dto.ShortenURLInput
		ctx   context.Context
	}

	tests := []struct {
		name    string
		wantErr bool
		setup   func(repo *mocks.Mockrepo, cache *mocks.Mockcache) args
	}{
		{
			name:    "sad case: failed to shorten url - empty string",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo, _ *mocks.Mockcache) args {
				return args{
					input: &dto.ShortenURLInput{URL: ""},
					ctx:   context.Background(),
				}
			},
		},
		{
			name:    "sad case: failed to shorten url-missing scheme",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo, _ *mocks.Mockcache) args {
				return args{
					input: &dto.ShortenURLInput{URL: "www.example.com"},
					ctx:   context.Background(),
				}
			},
		},
		{
			name:    "sad case: failed to shorten url-missing host",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo, _ *mocks.Mockcache) args {
				return args{
					input: &dto.ShortenURLInput{URL: "https://"},
					ctx:   context.Background(),
				}
			},
		},
		{
			name:    "sad case: failed to shorten url-length exceeds maximum limit",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo, _ *mocks.Mockcache) args {
				return args{
					input: &dto.ShortenURLInput{URL: "https://example.com/data?query=" + sb.String()},
					ctx:   context.Background(),
				}
			},
		},
		{
			name:    "sad case: failed to shorten url-error checking for existence",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockcache) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("an error occurred"))

				return args{
					input: &dto.ShortenURLInput{URL: "http://example.com"},
					ctx:   helpers.SetUserIDCtx(context.Background(), dto.AnonymousUserID),
				}
			},
		},
		{
			name:    "sad case: failed to shorten url-error saving short url to db",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockcache) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, nil)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(nil, fmt.Errorf("an error occurred"))

				return args{
					input: &dto.ShortenURLInput{URL: "https://example.com"},
					ctx:   helpers.SetUserIDCtx(context.Background(), dto.AnonymousUserID),
				}
			},
		},
		{
			name:    "sad case: generated short code already exists",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockcache) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(&domain.Link{}, nil).Times(5)

				return args{
					input: &dto.ShortenURLInput{URL: "http://example.com"},
					ctx:   helpers.SetUserIDCtx(context.Background(), dto.AnonymousUserID),
				}
			},
		},
		{
			name:    "sad case: context missing user information",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo, _ *mocks.Mockcache) args {
				return args{
					input: &dto.ShortenURLInput{URL: "http://example.com"},
					ctx:   context.Background(),
				}
			},
		},
		{
			name:    "happy case: successfully shorten url for anonymous user",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockcache) args {
				token := helpers.HashSecret("")

				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{
					OwnershipToken: &token,
				}, nil)

				return args{
					input: &dto.ShortenURLInput{URL: "http://example.com"},
					ctx:   helpers.SetUserIDCtx(context.Background(), dto.AnonymousUserID),
				}
			},
		},
		{
			name:    "happy case: successfully shorten url with url path for authenticated user",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockcache) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{}, nil)

				return args{
					input: &dto.ShortenURLInput{URL: "https://www.example.com:8000/users?id=1&name=collins"},
					ctx:   helpers.SetUserIDCtx(context.Background(), gofakeit.UUID()),
				}
			},
		},
		{
			name:    "happy case: successfully shorten url with query params",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockcache) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{}, nil)

				return args{
					input: &dto.ShortenURLInput{
						URL:       "https://www.example.com?id=1&name=collins",
						ExpiresAt: time.Now().Add(2 * time.Hour),
					},
					ctx: helpers.SetUserIDCtx(context.Background(), gofakeit.UUID()),
				}
			},
		},
		{
			name:    "happy case: successfully shorten url with anchor",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockcache) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{}, nil)

				return args{
					input: &dto.ShortenURLInput{URL: "https://www.example.com?id=1&name=collins#home"},
					ctx:   helpers.SetUserIDCtx(context.Background(), gofakeit.UUID()),
				}
			},
		},
		{
			name:    "happy case: successfully shorten url with port number",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockcache) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{}, nil)

				return args{
					input: &dto.ShortenURLInput{URL: "https://www.example.com:8000?id=1&name=collins"},
					ctx:   helpers.SetUserIDCtx(context.Background(), gofakeit.UUID()),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockrepo(t)
			cache := mocks.NewMockcache(t)
			s := New(repo, cache)
			args := tt.setup(repo, cache)

			got, gotErr := s.ShortenURL(args.ctx, args.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("EncodeURL() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if got != nil {
				require.NotEmpty(t, got.ShortURL)
			}

			if tt.name == "happy case: successfully shorten url for anonymous user" {
				require.NotEmpty(t, got.OwnershipToken)
			}

			if tt.name == "happy case: successfully shorten url with url path for authenticated user" {
				require.Empty(t, got.OwnershipToken)
			}
		})
	}
}
