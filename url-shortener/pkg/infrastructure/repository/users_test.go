package repository

import (
	"context"
	"reflect"
	"testing"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
)

func TestRepository_CreateUser(t *testing.T) {
	tests := []struct {
		name    string
		input   *domain.User
		want    *domain.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New()
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.CreateUser(context.Background(), tt.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateUser() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			// TODO: update the condition below to compare got with tt.want.
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetUserByEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		want    *domain.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New()
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.GetUserByEmail(context.Background(), tt.email)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetUserByEmail() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserByEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetUserByOAuthID(t *testing.T) {
	tests := []struct {
		name            string
		oauthProvider   string
		oauthProviderID string
		want            *domain.User
		wantErr         bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New()
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.GetUserByOAuthID(context.Background(), tt.oauthProvider, tt.oauthProviderID)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetUserByOAuthID() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.wantErr) {
				t.Errorf("GetUserByOAuthID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetUserByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		want    *domain.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New()
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.GetUserByID(context.Background(), tt.id)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetUserByID() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetUserByPublicID(t *testing.T) {
	tests := []struct {
		name     string
		publicID string
		want     *domain.User
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New()
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.GetUserByPublicID(context.Background(), tt.publicID)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetUserByPublicID() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserByPublicID() = %v, want %v", got, tt.want)
			}
		})
	}
}
