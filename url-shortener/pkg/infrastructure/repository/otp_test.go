package repository

import (
	"context"
	"reflect"
	"testing"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
)

func TestRepository_CreateOTP(t *testing.T) {
	tests := []struct {
		name    string
		input   *domain.OTP
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

			gotErr := r.CreateOTP(context.Background(), tt.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateOTP() err = %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_GetOTPByCodeAndUser(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		user    string
		purpose dto.OTPPurpose
		want    *domain.OTP
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

			got, gotErr := r.GetOTPByCodeAndUser(context.Background(), tt.code, tt.user, tt.purpose)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetOTPByCodeAndUser() err = %v wantErr %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetOTPByCodeAndUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_RevokeAllOTPs(t *testing.T) {
	tests := []struct {
		name    string
		user    string
		purpose string
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

			gotErr := r.RevokeAllOTPsForUser(context.Background(), tt.user, tt.purpose)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("RevokeAllOTPs() err %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}
