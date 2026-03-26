package otp

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/otp/mocks"
	"github.com/stretchr/testify/mock"
)

var errMsg = errors.New("an error occurred")

func TestProvider_GenerateOTP(t *testing.T) {
	type args struct {
		ctx     context.Context
		userID  string
		purpose dto.OTPPurpose
	}

	tests := []struct {
		name    string
		wantErr bool
		setup   func(repo *mocks.Mockrepository) args
	}{
		{
			name:    "happy case: successfully create otp",
			wantErr: false,
			setup: func(repo *mocks.Mockrepository) args {
				ctx := context.Background()

				repo.EXPECT().CreateOTP(mock.Anything, mock.AnythingOfType("*domain.OTP")).Return(nil)

				return args{ctx: ctx, userID: gofakeit.UUID(), purpose: dto.Login}
			},
		},
		{
			name:    "sad case: create otp failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepository) args {
				ctx := context.Background()

				repo.EXPECT().CreateOTP(mock.Anything, mock.AnythingOfType("*domain.OTP")).Return(errMsg)

				return args{ctx: ctx, userID: gofakeit.UUID(), purpose: dto.Login}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockrepository(t)
			p := New(repo)
			args := tt.setup(repo)

			_, gotErr := p.GenerateOTP(args.ctx, args.userID, args.purpose)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GenerateOTP() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestProvider_GetOTPByCodeAndUserID(t *testing.T) {
	type args struct {
		code, userID string
		purpose      dto.OTPPurpose
	}

	tests := []struct {
		name    string
		wantErr bool
		setup   func(repo *mocks.Mockrepository) args
	}{
		{
			name:    "happy case: get otp successful",
			wantErr: false,
			setup: func(repo *mocks.Mockrepository) args {
				userID := gofakeit.UUID()
				purpose := dto.Login

				repo.EXPECT().GetOTPByCodeAndUser(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&domain.OTP{
					PublicID: userID,
					Code:     "000000",
					Purpose:  purpose,
				}, nil)

				return args{code: "000000", userID: userID, purpose: purpose}
			},
		},
		{
			name:    "sad case: get otp failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepository) args {
				userID := gofakeit.UUID()
				purpose := dto.Login

				repo.EXPECT().GetOTPByCodeAndUser(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errMsg)

				return args{code: "000000", userID: userID, purpose: purpose}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockrepository(t)
			p := New(repo)
			args := tt.setup(repo)

			_, gotErr := p.GetOTPByCodeAndUserID(context.Background(), args.code, args.userID, args.purpose)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetOTPByCodeAndUserID() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestProvider_RevokeAllOTPsForUser(t *testing.T) {
	type args struct {
		user    string
		purpose dto.OTPPurpose
	}

	tests := []struct {
		name    string
		setup   func(repo *mocks.Mockrepository) args
		wantErr bool
	}{
		{
			name:    "happy case: revoke all otps for user successful",
			wantErr: false,
			setup: func(repo *mocks.Mockrepository) args {
				repo.EXPECT().RevokeAllOTPsForUser(mock.Anything, mock.Anything, mock.Anything).Return(nil)

				return args{user: gofakeit.UUID(), purpose: dto.Login}
			},
		},
		{
			name:    "sad case: revoke all otps for user failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepository) args {
				repo.EXPECT().RevokeAllOTPsForUser(mock.Anything, mock.Anything, mock.Anything).Return(errMsg)

				return args{user: gofakeit.UUID(), purpose: dto.Login}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockrepository(t)
			p := New(repo)
			args := tt.setup(repo)

			gotErr := p.RevokeAllOTPsForUser(context.Background(), args.user, args.purpose)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("RevokeAllOTPsForUser() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}
