package repository

import (
	"testing"
	"time"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/jxskiss/base62"
)

func TestRepository_CreateShortLink(t *testing.T) {
	url := "http://example.com"
	now := time.Now()
	expiry := now.Add(1 * time.Millisecond)
	token := "dfcbd7c92a553992e5c1df63c71e979089d94c7bbcec63e2f4649dcb1e292e7b"

	type args struct {
		input domain.Link
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy case: create short link successful - no expiry",
			args: args{
				input: domain.Link{
					UserID:         1,
					ShortCode:      base62.EncodeToString([]byte(url)),
					OriginalURL:    url,
					OwnershipToken: token,
					CreatedAt:      now,
				},
			},
			wantErr: false,
		},
		{
			name: "happy case: create short link successful - with expiry",
			args: args{
				input: domain.Link{
					UserID:         1,
					ShortCode:      base62.EncodeToString([]byte(url)),
					OriginalURL:    url,
					OwnershipToken: token,
					CreatedAt:      now,
					ExpiresAt:      &expiry,
				},
			},
			wantErr: false,
		},
		{
			name: "sad case: create short link failed",
			args: args{
				input: domain.Link{
					UserID:         3,
					ShortCode:      base62.EncodeToString([]byte(url)),
					OriginalURL:    url,
					OwnershipToken: token,
					CreatedAt:      now,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.CreateShortLink(t.Context(), tt.args.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateShortLink() err = %v wantErr: %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_GetLinkByCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "happy case: successfully get link by code",
			code:    base62.EncodeToString([]byte("http://example.com")),
			wantErr: false,
		},
		{
			name:    "sad case: error getting link by code",
			code:    base62.EncodeToString([]byte("http://unknown-url.com")),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.GetLinkByCode(t.Context(), tt.code)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetLinkByCode() err =  %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}
