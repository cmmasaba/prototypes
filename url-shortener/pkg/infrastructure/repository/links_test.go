package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/jxskiss/base62"
)

func TestRepository_CreateShortLink(t *testing.T) {
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
			r, err := New()
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.CreateShortLink(tt.args.ctx, tt.args.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateShortLink() err = %v wantErr: %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateShortLink() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetLinkByCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		want    *domain.Link
		wantErr bool
	}{
		{
			name:    "happy case: successfully get link by code",
			code:    "",
			want:    &domain.Link{},
			wantErr: false,
		},
		{
			name:    "sad case: error getting link by code",
			code:    "",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New()
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.GetLinkByCode(context.Background(), tt.code)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetLinkByCode() err =  %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLinkByCode() got = %v, want = %v", got, tt.want)
			}
		})
	}
}
