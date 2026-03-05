package repository

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/cmmasaba/prototypes/pkg/application/domain"
	"github.com/jxskiss/base62"
)

func TestRepository_SaveShortLink(t *testing.T) {
	postgresURL, ok := os.LookupEnv("POSTGRES_URL")
	if !ok {
		t.Fatal("mandatory environment variable POSTGRES_URL not set")
	}

	url := gofakeit.URL()

	validLink := domain.Link{
		ShortCode:      base62.EncodeToString([]byte(url)),
		OriginalURL:    url,
		OwnershipToken: gofakeit.BeerName(),
		CreatedAt:      time.Now(),
	}

	type args struct {
		ctx   context.Context
		input domain.Link
	}

	tests := []struct {
		name    string
		args    args
		want    *domain.Link
		wantErr bool
	}{
		{
			name: "happy case: successfully save short link",
			args: args{
				ctx:   context.Background(),
				input: validLink,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRepository(postgresURL)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.SaveShortLink(tt.args.ctx, tt.args.input)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("SaveShortLink() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("SaveShortLink() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SaveShortLink() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetLinkByCode(t *testing.T) {
	postgresURL, ok := os.LookupEnv("POSTGRES_URL")
	if !ok {
		t.Fatal("mandatory environment variable POSTGRES_URL not set")
	}

	tests := []struct {
		name       string
		connString string
		code       string
		want       *domain.Link
		wantErr    bool
	}{
		{
			name:       "happy case: successfully get link by code",
			connString: postgresURL,
			code:       "",
			want:       &domain.Link{},
			wantErr:    false,
		},
		{
			name:       "sad case: error getting link by code",
			connString: postgresURL,
			code:       "",
			want:       nil,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRepository(tt.connString)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.GetLinkByCode(context.Background(), tt.code)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetLinkByCode() failed: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("GetLinkByCode() succeeded unexpectedly")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLinkByCode() got = %v, want = %v", got, tt.want)
			}
		})
	}
}
