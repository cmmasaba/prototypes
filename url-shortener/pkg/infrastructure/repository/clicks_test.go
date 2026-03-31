package repository

import (
	"testing"
	"time"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
)

func TestRepository_CreateClick(t *testing.T) {
	county := "KE"

	tests := []struct {
		name    string
		input   domain.Click
		wantErr bool
	}{
		{
			name: "happy case: create click successfully",
			input: domain.Click{
				LinkID:    1,
				ClickedAt: time.Date(2025, 0o2, 15, 14, 45, 45, 78, time.UTC),
				Country:   &county,
			},
			wantErr: false,
		},
		{
			name: "sad case: create click failed",
			input: domain.Click{
				LinkID:    100,
				ClickedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.CreateClick(t.Context(), tt.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateClickData() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_GetClicksByLinkIDAndClickedAt(t *testing.T) {
	tests := []struct {
		name      string
		linkID    int64
		clickedAt time.Time
		wantErr   bool
	}{
		{
			name:      "happy case: get click by link id and clickedat successfully",
			linkID:    1,
			clickedAt: time.Date(2025, 0o2, 15, 14, 45, 45, 78, time.UTC),
			wantErr:   false,
		},
		{
			name:      "sad case: get click by link id and clickedat failed",
			linkID:    100,
			clickedAt: time.Now(),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.GetClicksByLinkIDAndClickedAt(t.Context(), tt.linkID, tt.clickedAt)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetClicksByLinkIDAndClickedAt() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_GetClicksByLinkIDAndCountry(t *testing.T) {
	country := "KE"

	tests := []struct {
		name    string
		linkID  int64
		country *string
		wantErr bool
	}{
		{
			name:    "happy case: get click by ink id and country successful",
			linkID:  1,
			country: &country,
			wantErr: false,
		},
		{
			name:    "sad case: get click by ink id and country failed",
			linkID:  100,
			country: &country,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.GetClicksByLinkIDAndCountry(t.Context(), tt.linkID, tt.country)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetClicksByLinkIDAndCountry() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}
