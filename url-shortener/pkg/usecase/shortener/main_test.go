package shortener

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/shortener/mocks"
	"github.com/stretchr/testify/mock"
)

func TestShortener_ShortenURL(t *testing.T) {
	var sb strings.Builder
	for range 2500 {
		sb.WriteString("a")
	}

	type args struct {
		url string
	}

	tests := []struct {
		name    string
		wantErr bool
		setup   func(repo *mocks.Mockrepo) args
	}{
		{
			name:    "sad case: failed to shorten url - empty string",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo) args {
				return args{url: ""}
			},
		},
		{
			name:    "sad case: failed to shorten url-missing scheme",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo) args {
				return args{url: "www.example.com"}
			},
		},
		{
			name:    "sad case: failed to shorten url-missing host",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo) args {
				return args{url: "https://"}
			},
		},
		{
			name:    "sad case: failed to shorten url-length exceeds maximum limit",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo) args {
				return args{url: "https://example.com/data?query=" + sb.String()}
			},
		},
		{
			name:    "sad case: failed to shorten url-error checking for existence",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("an error occurred"))

				return args{url: "http://example.com"}
			},
		},
		{
			name:    "sad case: failed to shorten url-error saving short url to db",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, nil)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(nil, fmt.Errorf("an error occurred"))

				return args{url: "http://example.com"}
			},
		},
		{
			name:    "sad case: generated short code already exists",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(&domain.Link{}, nil).Times(5)

				return args{url: "http://example.com"}
			},
		},
		{
			name:    "happy case: successfully shorten url",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{}, nil)

				return args{url: "http://example.com"}
			},
		},
		{
			name:    "happy case: successfully shorten url with query params",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{}, nil)

				return args{url: "https://www.example.com?id=1&name=collins"}
			},
		},
		{
			name:    "happy case: successfully shorten url with anchor",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{}, nil)

				return args{url: "https://www.example.com?id=1&name=collins#home"}
			},
		},
		{
			name:    "happy case: successfully shorten url with port number",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{}, nil)

				return args{url: "https://www.example.com:8080?id=1&name=collins#home"}
			},
		},
		{
			name:    "happy case: successfully shorten url with url path",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo) args {
				repo.EXPECT().GetLinkByCode(mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
				repo.EXPECT().CreateShortLink(mock.Anything, mock.AnythingOfType("domain.Link")).Return(&domain.Link{}, nil)

				return args{url: "https://www.example.com:8080/users?id=1&name=collins#home"}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockrepo(t)
			s := New(repo)
			args := tt.setup(repo)

			_, gotErr := s.ShortenURL(t.Context(), args.url)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("EncodeURL() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}
