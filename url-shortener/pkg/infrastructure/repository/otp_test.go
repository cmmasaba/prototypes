package repository

import (
	"testing"
	"time"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
)

func TestRepository_CreateOTP(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		input   *domain.OTP
		wantErr bool
	}{
		{
			name: "happy case: create otp successful",
			input: &domain.OTP{
				User: domain.User{
					PublicID: "019d402e-4075-731a-ba42-0432d5ced9",
				},
				Code:      "d1f7f5986d849521ca17b46d834c3c2f5ecf611047640ed69b6e1859894b366b",
				ExpiresAt: now.Add(15 * time.Minute),
				Revoked:   false,
				Purpose:   dto.Login,
			},
			wantErr: false,
		},
		{
			name: "sad case: create otp failed - unknown purpose",
			input: &domain.OTP{
				User: domain.User{
					PublicID: "019d402e-4075-731a-ba42-0432d5ced9",
				},
				Code:      "d1f7f5986d849521ca17b46d834c3c2f5ecf611047640ed69b6e1859894b366b",
				ExpiresAt: now.Add(15 * time.Minute),
				Revoked:   false,
				Purpose:   dto.OTPPurpose("INVALID"),
			},
			wantErr: true,
		},
		{
			name: "sad case: create otp failed - invalid public id",
			input: &domain.OTP{
				User: domain.User{
					PublicID: "019d402e-4075-731a-ba42-0432d5ced9",
				},
				Code:      "d1f7f5986d849521ca17b46d834c3c2f5ecf611047640ed69b6e1859894b366b",
				ExpiresAt: now.Add(15 * time.Minute),
				Revoked:   false,
				Purpose:   dto.Login,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := testRepository.CreateOTP(t.Context(), tt.input)
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
		wantErr bool
	}{
		{
			name:    "happy case: get otp by code and user successful",
			code:    "d1f7f5986d849521ca17b46d834c3c2f5ecf611047640ed69b6e1859894b366b",
			user:    "019d402e-4075-731a-ba42-0432d5cdef99",
			purpose: dto.Login,
			wantErr: false,
		},
		{
			name:    "sad case: get otp by code and user successful - invalid uuid",
			code:    "d1f7f5986d849521ca17b46d834c3c2f5ecf611047640ed69b6e1859894b366b",
			user:    "019d402e-4075-731a-ba42-0432d5cdef",
			purpose: dto.Login,
			wantErr: true,
		},
		{
			name:    "sad case: get otp by code and user failed - missing purpose",
			code:    "d1f7f5986d849521ca17b46d834c3c2f5ecf611047640ed69b6e1859894b366b",
			user:    "019d402e-4075-731a-ba42-0432d5cdef99",
			purpose: dto.EmailVerification,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := testRepository.GetOTPByCodeAndUser(t.Context(), tt.code, tt.user, tt.purpose)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetOTPByCodeAndUser() err = %v wantErr %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestRepository_RevokeAllOTPsForUser(t *testing.T) {
	tests := []struct {
		name    string
		user    string
		purpose string
		wantErr bool
	}{
		{
			name:    "happy case: revoke all otps for user successful",
			user:    "019d402e-4075-731a-ba42-0432d5cdef99",
			purpose: dto.Login.String(),
			wantErr: false,
		},
		{
			name:    "sad case: revoke all otps for user failed",
			user:    "019d402e-4075-731a-ba42-0432d5cde",
			purpose: dto.EmailVerification.String(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := testRepository.RevokeAllOTPsForUser(t.Context(), tt.user, tt.purpose)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("RevokeAllOTPs() err %v wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}
