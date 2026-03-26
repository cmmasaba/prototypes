package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
)

func TestRepository_GetClicksByLinkIDAndClickedAt(t *testing.T) {
	tests := []struct {
		name      string
		linkID    int64
		clickedAt time.Time
		want      []*domain.Click
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New()
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}

			got, gotErr := r.GetClicksByLinkIDAndClickedAt(context.Background(), tt.linkID, tt.clickedAt)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetClicksByLinkIDAndClickedAt() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClicksByLinkIDAndClickedAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_CreateClickData(t *testing.T) {
	tests := []struct {
		name    string
		data    domain.Click
		want    *domain.Click
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

			got, gotErr := r.CreateClickData(context.Background(), tt.data)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateClickData() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateClickData() = %v, want %v", got, tt.want)

				return
			}
		})
	}
}

func TestRepository_GetClicksByLinkIDAndCountry(t *testing.T) {
	tests := []struct {
		name    string
		linkID  int64
		country *string
		want    []*domain.Click
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

			got, gotErr := r.GetClicksByLinkIDAndCountry(context.Background(), tt.linkID, tt.country)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetClicksByLinkIDAndCountry() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClicksByLinkIDAndCountry() = %v, want %v", got, tt.want)

				return
			}
		})
	}
}
