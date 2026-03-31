package repository

import (
	"testing"
	"time"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
)

func TestRepository_SaveRefreshToken(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		input   domain.RefreshToken
		wantErr bool
	}{
		{
			name: "happy case: save refresh token successful",
			input: domain.RefreshToken{
				UserID:    1,
				Token:     "dfcbd7c92a553992e5c1df63c71e979089d94c7bbcec63e2f4649dcb1e292e7b",
				Revoked:   false,
				CreatedAt: now,
				ExpiresAt: now.Add(15 * time.Minute),
			},
			wantErr: false,
		},
		{
			name: "sad case: save refresh token failed",
			input: domain.RefreshToken{
				UserID:    100,
				Token:     "dfcbd7c92a553992e5c1df63c71e979089d94c7bbcec63e2f4649dcb1e293d5g",
				Revoked:   false,
				CreatedAt: now,
				ExpiresAt: now.Add(15 * time.Minute),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := testRepository.SaveRefreshToken(t.Context(), tt.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("SaveRefreshToken() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_GetRefreshTokenByTokenHash(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "happy case: get refresh token by token hash successful",
			token:   "dfcbd7c92a553992e5c1df63c71e979089d94c7bbcec63e2f4649dcb1e292e7b",
			wantErr: false,
		},
		{
			name:    "sad case: get refresh token by token hash failed",
			token:   "dfcbd7c92a553992e5c1df63c71e979089d94c7bbcec63e2f4649dcb1e292b7e",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.GetRefreshTokenByTokenHash(t.Context(), tt.token)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetRefreshTokenByTokenHash() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_RevokeRefreshToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "happy case: revoke refresh token successful",
			token:   "dfcbd7c92a553992e5c1df63c71e979089d94c7bbcec63e2f4649dcb1e292e7b",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := testRepository.RevokeRefreshToken(t.Context(), tt.token)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("RevokeRefreshToken() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}
