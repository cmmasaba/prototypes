package repository

import (
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
)

func TestRepository_CreateUser(t *testing.T) {
	user1Password := "$2a$14$nQkSh1bFGG3wTc8sC0EHNOezB3CDFKca9woCT77UKeEVqKv1kCmue"
	user2Password := "$2a$14$nQkSh1bFGG3wTc8sC0EHNOezB3CDFKca9woCT77UKeEVqKv1kDuem"
	email := gofakeit.Email()

	tests := []struct {
		name    string
		input   *domain.User
		wantErr bool
	}{
		{
			name: "happy case: create user successfully",
			input: &domain.User{
				Email:        email,
				PasswordHash: &user1Password,
			},
			wantErr: false,
		},
		{
			name: "sad case: create user failed - duplicate email",
			input: &domain.User{
				Email:        email,
				PasswordHash: &user2Password,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.CreateUser(t.Context(), tt.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateUser() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "happy case: get user by email successful",
			email:   "test@email.com",
			wantErr: false,
		},
		{
			name:    "sad case: get user by email failed",
			email:   gofakeit.Email(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.GetUserByEmail(t.Context(), tt.email)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetUserByEmail() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_GetUserByOAuthProviderAndID(t *testing.T) {
	knownID := "6aa8ecb7-51ad-4287-aa81-1cc738d9a81a"
	unknownID := "2393ff66-3a97-41af-8d8c-559161313ee7"

	tests := []struct {
		name            string
		oauthProvider   string
		oauthProviderID string
		wantErr         bool
	}{
		{
			name:            "happy case: get user by oauthID successful",
			oauthProvider:   "Google",
			oauthProviderID: knownID,
			wantErr:         false,
		},
		{
			name:            "sad case: get user by oauthID failed",
			oauthProvider:   "Google",
			oauthProviderID: unknownID,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.GetUserByOAuthProviderAndID(t.Context(), tt.oauthProvider, tt.oauthProviderID)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetUserByOAuthID() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_GetUserByID(t *testing.T) {
	knownID := 1
	unknownID := 5000

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{
			name:    "happy case: get user by is successful",
			id:      int64(knownID),
			wantErr: false,
		},
		{
			name:    "sad case: get user by id failed",
			id:      int64(unknownID),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.GetUserByID(t.Context(), tt.id)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetUserByID() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_GetUserByPublicID(t *testing.T) {
	knownID := "019d400e-74a2-7e3e-90a9-4761dae23795"
	unknownID := "019d4010-2834-7e1c-acff-6d0ce98ef648"

	tests := []struct {
		name     string
		publicID string
		wantErr  bool
	}{
		{
			name:     "happy case: get user by publicID successful",
			publicID: knownID,
			wantErr:  false,
		},
		{
			name:     "sad case: get user by publicID failed",
			publicID: unknownID,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.GetUserByPublicID(t.Context(), tt.publicID)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetUserByPublicID() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}
