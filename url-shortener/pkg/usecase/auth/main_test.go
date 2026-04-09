package auth

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
)

var errMsg = errors.New("an error occurred")

func TestUsecaseImpl_CreateUserEmailPassword(t *testing.T) {
	responseStatus := "otp_required"

	type args struct {
		ctx   context.Context
		input *dto.EmailPasswordUserInput
		want  *dto.OTPRequiredResponse
	}

	tests := []struct {
		name    string
		wantErr bool
		setup   func(
			repo *mocks.Mockrepo,
			otp *mocks.Mockotp,
			tasks *mocks.MockbackgroundTasks,
			hibp *mocks.Mockhibp,
		) args
	}{
		{
			name: "happy case: successfully create new user",
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp, tasks *mocks.MockbackgroundTasks, hibp *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)
				resp := &domain.User{
					PublicID:  gofakeit.UUID(),
					Email:     email,
					CreatedAt: time.Now(),
				}

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(nil, repository.ErrNotFound)
				hibp.EXPECT().CheckPasswordIsBreached(mock.Anything, password).Return(false, nil)
				repo.EXPECT().CreateUser(mock.Anything, mock.AnythingOfType("*domain.User")).Return(resp, nil)
				otp.EXPECT().GenerateOTP(mock.Anything, mock.Anything, dto.EmailVerification).Return("000000", nil)
				tasks.EXPECT().
					NewEmailDeliveryTask(mock.Anything, mock.AnythingOfType("tasks.EmailDeliveryPayload"), mock.AnythingOfType("tasks.Priority")).
					Return(nil)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want: &dto.OTPRequiredResponse{
						User:   dto.User{PublicID: resp.PublicID, Email: resp.Email, CreatedAt: resp.CreatedAt},
						Status: responseStatus,
					},
				}
			},
			wantErr: false,
		},
		{
			name: "happy case: create user for existing email",
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockotp, tasks *mocks.MockbackgroundTasks, _ *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(&domain.User{Email: email}, nil)
				tasks.EXPECT().
					NewEmailDeliveryTask(mock.Anything, mock.AnythingOfType("tasks.EmailDeliveryPayload"), mock.AnythingOfType("tasks.Priority")).
					Return(nil)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want: &dto.OTPRequiredResponse{
						User:   dto.User{Email: email},
						Status: responseStatus,
					},
				}
			},
			wantErr: false,
		},
		{
			name: "sad case: bad email format",
			setup: func(_ *mocks.Mockrepo, _ *mocks.Mockotp, _ *mocks.MockbackgroundTasks, _ *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.BeerName()
				password := gofakeit.Password(true, true, true, true, false, 10)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want:  nil,
				}
			},
			wantErr: true,
		},
		{
			name: "sad case: get user by email failed",
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockotp, _ *mocks.MockbackgroundTasks, _ *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(nil, errMsg)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want:  nil,
				}
			},
			wantErr: true,
		},
		{
			name: "sad case: send email for existing user failed",
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockotp, tasks *mocks.MockbackgroundTasks, _ *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(&domain.User{Email: email}, nil)
				tasks.EXPECT().
					NewEmailDeliveryTask(mock.Anything, mock.AnythingOfType("tasks.EmailDeliveryPayload"), mock.AnythingOfType("tasks.Priority")).
					Return(errMsg)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want: &dto.OTPRequiredResponse{
						User:   dto.User{Email: email},
						Status: responseStatus,
					},
				}
			},
			wantErr: false,
		},
		{
			name: "sad case: check password breach failed",
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp, tasks *mocks.MockbackgroundTasks, hibp *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)
				resp := &domain.User{
					PublicID:  gofakeit.UUID(),
					Email:     email,
					CreatedAt: time.Now(),
				}

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(nil, repository.ErrNotFound)
				hibp.EXPECT().CheckPasswordIsBreached(mock.Anything, password).Return(false, errMsg)
				repo.EXPECT().CreateUser(mock.Anything, mock.AnythingOfType("*domain.User")).Return(resp, nil)
				otp.EXPECT().GenerateOTP(mock.Anything, mock.Anything, dto.EmailVerification).Return("000000", nil)
				tasks.EXPECT().
					NewEmailDeliveryTask(mock.Anything, mock.AnythingOfType("tasks.EmailDeliveryPayload"), mock.AnythingOfType("tasks.Priority")).
					Return(nil)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want: &dto.OTPRequiredResponse{
						User:   dto.User{PublicID: resp.PublicID, Email: resp.Email, CreatedAt: resp.CreatedAt},
						Status: responseStatus,
					},
				}
			},
			wantErr: false,
		},
		{
			name: "sad case: password is breached",
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockotp, _ *mocks.MockbackgroundTasks, hibp *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(nil, repository.ErrNotFound)
				hibp.EXPECT().CheckPasswordIsBreached(mock.Anything, password).Return(true, nil)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want:  nil,
				}
			},
			wantErr: true,
		},
		{
			name: "sad case: save user to db failed",
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockotp, _ *mocks.MockbackgroundTasks, hibp *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(nil, repository.ErrNotFound)
				hibp.EXPECT().CheckPasswordIsBreached(mock.Anything, password).Return(false, errMsg)
				repo.EXPECT().CreateUser(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil, errMsg)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want:  nil,
				}
			},
			wantErr: true,
		},
		{
			name: "sad case: send email verification otp failed",
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp, tasks *mocks.MockbackgroundTasks, hibp *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)
				resp := &domain.User{
					PublicID:  gofakeit.UUID(),
					Email:     email,
					CreatedAt: time.Now(),
				}

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(nil, repository.ErrNotFound)
				hibp.EXPECT().CheckPasswordIsBreached(mock.Anything, password).Return(false, nil)
				repo.EXPECT().CreateUser(mock.Anything, mock.AnythingOfType("*domain.User")).Return(resp, nil)
				otp.EXPECT().GenerateOTP(mock.Anything, mock.Anything, dto.EmailVerification).Return("000000", nil)
				tasks.EXPECT().
					NewEmailDeliveryTask(mock.Anything, mock.AnythingOfType("tasks.EmailDeliveryPayload"), mock.AnythingOfType("tasks.Priority")).
					Return(errMsg)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want:  nil,
				}
			},
			wantErr: true,
		},
		{
			name: "sad case: generate otp failed",
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp, _ *mocks.MockbackgroundTasks, hibp *mocks.Mockhibp) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)
				resp := &domain.User{
					PublicID:  gofakeit.UUID(),
					Email:     email,
					CreatedAt: time.Now(),
				}

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(nil, repository.ErrNotFound)
				hibp.EXPECT().CheckPasswordIsBreached(mock.Anything, password).Return(false, nil)
				repo.EXPECT().CreateUser(mock.Anything, mock.AnythingOfType("*domain.User")).Return(resp, nil)
				otp.EXPECT().GenerateOTP(mock.Anything, mock.Anything, dto.EmailVerification).Return("", errMsg)

				return args{
					ctx:   ctx,
					input: &dto.EmailPasswordUserInput{Email: email, Password: password},
					want:  nil,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, hibp, otp, tasks, cache := mocks.NewMockrepo(
				t,
			), mocks.NewMockhibp(
				t,
			), mocks.NewMockotp(
				t,
			), mocks.NewMockbackgroundTasks(
				t,
			), mocks.NewMockcache(
				t,
			)
			u := New(repo, hibp, otp, tasks, cache)
			args := tt.setup(repo, otp, tasks, hibp)

			got, gotErr := u.CreateUserEmailPassword(context.Background(), args.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("CreateUserEmailPassword() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, args.want) {
				t.Errorf("CreateUserEmailPassword() = %v, want %v", got, args.want)

				return
			}
		})
	}
}

func TestUsecaseImpl_ValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name  string
		input *dto.ValidatePasswordInput
		want  bool
	}{
		{
			name: "happy case: strong password",
			input: &dto.ValidatePasswordInput{
				Password: gofakeit.Password(true, true, true, true, false, 10),
			},
			want: true,
		},
		{
			name: "sad case: weak password",
			input: &dto.ValidatePasswordInput{
				Password: "abcd",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, hibp, otp, tasks, cache := mocks.NewMockrepo(
				t,
			), mocks.NewMockhibp(
				t,
			), mocks.NewMockotp(
				t,
			), mocks.NewMockbackgroundTasks(
				t,
			), mocks.NewMockcache(
				t,
			)
			u := New(repo, hibp, otp, tasks, cache)

			got := u.ValidatePasswordStrength(context.Background(), tt.input)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidatePasswordStrength() = %v, want %v", got, tt.want)

				return
			}
		})
	}
}

func TestUsecaseImpl_ValidateJWTToken(t *testing.T) {
	type args struct {
		ctx         context.Context
		tokenString string
		want        jwt.MapClaims
	}

	accessTokenTTL := 15 * time.Second
	accessKey := helpers.MustGetEnvVar("JWT_ACCESS_SIGNING_KEY")

	tests := []struct {
		name    string
		wantErr bool
		setup   func() args
	}{
		{
			name:    "happy case: successful jwt validation",
			wantErr: false,
			setup: func() args {
				ctx := context.Background()
				now := time.Now()
				subject := gofakeit.UUID()
				accessTokenClaims := jwt.MapClaims{
					"exp": now.Add(accessTokenTTL).Unix(),
					"iat": now.Unix(),
					"sub": subject,
					"iss": "url-shortener",
					"aud": "url-shortener-api",
				}
				accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessTokenClaims)

				accessTokenString, err := accessToken.SignedString([]byte(accessKey))
				if err != nil {
					t.Errorf("jwt access token signing failed: %v", err)
				}

				return args{
					ctx: ctx, tokenString: accessTokenString,
					want: jwt.MapClaims{
						"exp": float64(now.Add(accessTokenTTL).Unix()),
						"iat": float64(now.Unix()),
						"sub": subject,
						"iss": "url-shortener",
						"aud": "url-shortener-api",
					},
				}
			},
		},
		{
			name:    "sad case: failed jwt validation",
			wantErr: true,
			setup: func() args {
				ctx := context.Background()
				now := time.Now()
				accessTokenClaims := jwt.MapClaims{
					"exp": now.Add(accessTokenTTL).Unix(),
					"iat": now.Unix(),
					"sub": gofakeit.UUID(),
					"iss": "url-short",
					"aud": "url-short-api",
				}
				accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessTokenClaims)

				accessTokenString, err := accessToken.SignedString([]byte(accessKey))
				if err != nil {
					t.Errorf("jwt access token signing failed: %v", err)
				}

				return args{ctx: ctx, tokenString: accessTokenString, want: nil}
			},
		},
		{
			name:    "sad case: expired jwt token",
			wantErr: true,
			setup: func() args {
				ctx := context.Background()
				accessTokenClaims := jwt.MapClaims{
					"exp": time.Date(2026, 0o1, 0o1, 11, 59, 14, 67, time.UTC).Unix(),
					"iat": time.Date(2026, 0o1, 0o1, 11, 44, 14, 67, time.UTC).Unix(),
					"sub": gofakeit.UUID(),
					"iss": "url-shortener",
					"aud": "url-shortener-api",
				}
				accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, accessTokenClaims)

				accessTokenString, err := accessToken.SignedString([]byte(accessKey))
				if err != nil {
					t.Errorf("jwt access token signing failed: %v", err)
				}

				return args{ctx: ctx, tokenString: accessTokenString, want: nil}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, hibp, otp, tasks, cache := mocks.NewMockrepo(
				t,
			), mocks.NewMockhibp(
				t,
			), mocks.NewMockotp(
				t,
			), mocks.NewMockbackgroundTasks(
				t,
			), mocks.NewMockcache(
				t,
			)
			u := New(repo, hibp, otp, tasks, cache)

			args := tt.setup()

			got, gotErr := u.ValidateJWTToken(context.Background(), args.tokenString)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("ValidateJWTToken() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, args.want) {
				t.Errorf("ValidateJWTToken() = %v, want %v", got, args.want)

				return
			}
		})
	}
}

func TestUsecaseImpl_Login(t *testing.T) {
	type args struct {
		ctx   context.Context
		input *dto.LoginInput
		want  *dto.OTPRequiredResponse
	}

	tests := []struct {
		name    string
		wantErr bool
		setup   func(repo *mocks.Mockrepo, otp *mocks.Mockotp, tasks *mocks.MockbackgroundTasks) args
	}{
		{
			name:    "happy case: successfully login",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp, tasks *mocks.MockbackgroundTasks) args {
				ctx := context.Background()
				email := gofakeit.Email()
				publicID := gofakeit.UUID()
				createdAt := time.Now()
				password := gofakeit.Password(true, true, true, true, false, 10)
				hashedPassword, _ := hashPassword(ctx, password)

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(&domain.User{
					ID:           1,
					PublicID:     publicID,
					PasswordHash: &hashedPassword,
					Email:        email,
					CreatedAt:    createdAt,
				}, nil)
				otp.EXPECT().GenerateOTP(mock.Anything, mock.Anything, dto.Login).Return("000000", nil)
				tasks.EXPECT().
					NewEmailDeliveryTask(mock.Anything, mock.AnythingOfType("tasks.EmailDeliveryPayload"), mock.AnythingOfType("tasks.Priority")).
					Return(nil)

				return args{
					ctx:   ctx,
					input: &dto.LoginInput{Email: email, Password: password},
					want: &dto.OTPRequiredResponse{User: dto.User{
						Email:     email,
						PublicID:  publicID,
						CreatedAt: createdAt,
					}, Status: "otp_required"},
				}
			},
		},
		{
			name:    "sad case: get user by email failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockotp, _ *mocks.MockbackgroundTasks) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(nil, errMsg)

				return args{
					ctx:   ctx,
					input: &dto.LoginInput{Email: email, Password: password},
					want:  nil,
				}
			},
		},
		{
			name:    "sad case: user has no password setup",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockotp, _ *mocks.MockbackgroundTasks) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(&domain.User{
					ID:           1,
					PublicID:     gofakeit.UUID(),
					PasswordHash: nil,
				}, nil)

				return args{
					ctx:   ctx,
					input: &dto.LoginInput{Email: email, Password: password},
					want:  nil,
				}
			},
		},
		{
			name:    "sad case: password mismatch",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockotp, _ *mocks.MockbackgroundTasks) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)
				hashedPassword, _ := hashPassword(ctx, gofakeit.Password(true, true, true, false, false, 12))

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(&domain.User{
					ID:           1,
					PublicID:     gofakeit.UUID(),
					PasswordHash: &hashedPassword,
				}, nil)

				return args{
					ctx:   ctx,
					input: &dto.LoginInput{Email: email, Password: password},
					want:  nil,
				}
			},
		},
		{
			name:    "sad case: generate otp failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp, _ *mocks.MockbackgroundTasks) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)
				hashedPassword, _ := hashPassword(ctx, password)
				now := time.Now()

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(&domain.User{
					ID:           1,
					PublicID:     gofakeit.UUID(),
					Email:        email,
					PasswordHash: &hashedPassword,
					CreatedAt:    now,
				}, nil)
				otp.EXPECT().GenerateOTP(mock.Anything, mock.Anything, dto.Login).Return("", errMsg)

				return args{
					ctx:   ctx,
					input: &dto.LoginInput{Email: email, Password: password},
					want:  nil,
				}
			},
		},
		{
			name:    "sad case: send otp email failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp, tasks *mocks.MockbackgroundTasks) args {
				ctx := context.Background()
				email := gofakeit.Email()
				password := gofakeit.Password(true, true, true, true, false, 10)
				hashedPassword, _ := hashPassword(ctx, password)
				now := time.Now()

				repo.EXPECT().GetUserByEmail(mock.Anything, email).Return(&domain.User{
					ID:           1,
					PublicID:     gofakeit.UUID(),
					Email:        email,
					PasswordHash: &hashedPassword,
					CreatedAt:    now,
				}, nil)
				otp.EXPECT().GenerateOTP(mock.Anything, mock.Anything, dto.Login).Return("000000", nil)
				tasks.EXPECT().
					NewEmailDeliveryTask(mock.Anything, mock.AnythingOfType("tasks.EmailDeliveryPayload"), mock.AnythingOfType("tasks.Priority")).
					Return(errMsg)

				return args{
					ctx:   ctx,
					input: &dto.LoginInput{Email: email, Password: password},
					want:  nil,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, hibp, otp, tasks, cache := mocks.NewMockrepo(
				t,
			), mocks.NewMockhibp(
				t,
			), mocks.NewMockotp(
				t,
			), mocks.NewMockbackgroundTasks(
				t,
			), mocks.NewMockcache(
				t,
			)
			u := New(repo, hibp, otp, tasks, cache)
			args := tt.setup(repo, otp, tasks)

			got, gotErr := u.Login(args.ctx, args.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Login() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}

			if !reflect.DeepEqual(got, args.want) {
				t.Errorf("Login() = %v, want %v", got, args.want)

				return
			}
		})
	}
}

func TestUsecaseImpl_RefreshAccessToken(t *testing.T) {
	type args struct {
		ctx          context.Context
		refreshToken string
	}

	tests := []struct {
		name    string
		wantErr bool
		setup   func(repo *mocks.Mockrepo) args
	}{
		{
			name:    "happy case: successfully refresh access token",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo) args {
				ctx := context.Background()

				repo.EXPECT().GetRefreshTokenByTokenHash(mock.Anything, mock.Anything).Return(&domain.RefreshToken{
					UserID:    1,
					ExpiresAt: time.Now().Add(3 * time.Minute),
					Revoked:   false,
				}, nil)
				repo.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       1,
					PublicID: gofakeit.UUID(),
				}, nil)
				repo.EXPECT().RevokeRefreshToken(mock.Anything, mock.Anything).Return(nil)
				repo.EXPECT().SaveRefreshToken(mock.Anything, mock.AnythingOfType("domain.RefreshToken")).Return(nil)

				return args{ctx: ctx, refreshToken: gofakeit.Name()}
			},
		},
		{
			name:    "sad case: failed to get refresh token from db",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo) args {
				ctx := context.Background()

				repo.EXPECT().GetRefreshTokenByTokenHash(mock.Anything, mock.Anything).Return(nil, errMsg)

				return args{ctx: ctx, refreshToken: gofakeit.Name()}
			},
		},
		{
			name:    "sad case: revoked token",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo) args {
				ctx := context.Background()

				repo.EXPECT().GetRefreshTokenByTokenHash(mock.Anything, mock.Anything).Return(&domain.RefreshToken{
					UserID:    1,
					ExpiresAt: time.Now().Add(3 * time.Minute),
					Revoked:   true,
				}, nil)

				return args{ctx: ctx, refreshToken: gofakeit.Name()}
			},
		},
		{
			name:    "sad case: expired token",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo) args {
				ctx := context.Background()

				repo.EXPECT().GetRefreshTokenByTokenHash(mock.Anything, mock.Anything).Return(&domain.RefreshToken{
					UserID:    1,
					ExpiresAt: time.Date(2026, 0o1, 0o1, 11, 30, 23, 46, time.UTC),
					Revoked:   false,
				}, nil)

				return args{ctx: ctx, refreshToken: gofakeit.Name()}
			},
		},
		{
			name:    "sad case: get user by id failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo) args {
				ctx := context.Background()

				repo.EXPECT().GetRefreshTokenByTokenHash(mock.Anything, mock.Anything).Return(&domain.RefreshToken{
					UserID:    1,
					ExpiresAt: time.Now().Add(3 * time.Minute),
					Revoked:   false,
				}, nil)
				repo.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(nil, errMsg)

				return args{ctx: ctx, refreshToken: gofakeit.Name()}
			},
		},
		{
			name:    "sad case: revoke otp failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo) args {
				ctx := context.Background()

				repo.EXPECT().GetRefreshTokenByTokenHash(mock.Anything, mock.Anything).Return(&domain.RefreshToken{
					UserID:    1,
					ExpiresAt: time.Now().Add(3 * time.Minute),
					Revoked:   false,
				}, nil)
				repo.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       1,
					PublicID: gofakeit.UUID(),
				}, nil)
				repo.EXPECT().RevokeRefreshToken(mock.Anything, mock.Anything).Return(errMsg)

				return args{ctx: ctx, refreshToken: gofakeit.Name()}
			},
		},
		{
			name:    "sad case: save refresh token to db failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo) args {
				ctx := context.Background()

				repo.EXPECT().GetRefreshTokenByTokenHash(mock.Anything, mock.Anything).Return(&domain.RefreshToken{
					UserID:    1,
					ExpiresAt: time.Now().Add(3 * time.Minute),
					Revoked:   false,
				}, nil)
				repo.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       1,
					PublicID: gofakeit.UUID(),
				}, nil)
				repo.EXPECT().RevokeRefreshToken(mock.Anything, mock.Anything).Return(nil)
				repo.EXPECT().SaveRefreshToken(mock.Anything, mock.AnythingOfType("domain.RefreshToken")).Return(errMsg)

				return args{ctx: ctx, refreshToken: gofakeit.Name()}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, hibp, otp, tasks, cache := mocks.NewMockrepo(
				t,
			), mocks.NewMockhibp(
				t,
			), mocks.NewMockotp(
				t,
			), mocks.NewMockbackgroundTasks(
				t,
			), mocks.NewMockcache(
				t,
			)
			u := New(repo, hibp, otp, tasks, cache)
			args := tt.setup(repo)

			_, gotErr := u.RefreshAccessToken(args.ctx, args.refreshToken)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("RefreshAccessToken() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}

func TestUsecaseImpl_VerifyOTP(t *testing.T) {
	type args struct {
		ctx   context.Context
		input *dto.VerifyOTPInput
	}

	tests := []struct {
		name    string
		setup   func(repo *mocks.Mockrepo, otp *mocks.Mockotp) args
		wantErr bool
	}{
		{
			name:    "happy case: successfully verify otp",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp) args {
				email, publicID, userID := gofakeit.Email(), gofakeit.UUID(), 1
				ctx := helpers.SetUserIDCtx(context.Background(), publicID)
				code := "000000"

				repo.EXPECT().GetUserByPublicID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       int64(userID),
					Email:    email,
					PublicID: publicID,
				}, nil)
				otp.EXPECT().
					GetOTPByCodeAndUserID(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.OTP{
						PublicID:  publicID,
						Code:      code,
						Revoked:   false,
						ExpiresAt: time.Now().Add(5 * time.Minute),
					}, nil)
				otp.EXPECT().RevokeAllOTPsForUser(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				repo.EXPECT().SaveRefreshToken(mock.Anything, mock.AnythingOfType("domain.RefreshToken")).Return(nil)

				return args{ctx: ctx, input: &dto.VerifyOTPInput{Purpose: dto.Login, Value: "000000"}}
			},
		},
		{
			name:    "sad case: save refresh token to db failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp) args {
				email, publicID, userID := gofakeit.Email(), gofakeit.UUID(), 1
				ctx := helpers.SetUserIDCtx(context.Background(), publicID)
				code := "000000"

				repo.EXPECT().GetUserByPublicID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       int64(userID),
					Email:    email,
					PublicID: publicID,
				}, nil)
				otp.EXPECT().
					GetOTPByCodeAndUserID(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.OTP{
						PublicID:  publicID,
						Code:      code,
						Revoked:   false,
						ExpiresAt: time.Now().Add(5 * time.Minute),
					}, nil)
				otp.EXPECT().RevokeAllOTPsForUser(mock.Anything, mock.Anything, mock.Anything).Return(nil)
				repo.EXPECT().SaveRefreshToken(mock.Anything, mock.AnythingOfType("domain.RefreshToken")).Return(errMsg)

				return args{ctx: ctx, input: &dto.VerifyOTPInput{Purpose: dto.Login, Value: "000000"}}
			},
		},
		{
			name:    "sad case: revoke all otps for user failed",
			wantErr: false,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp) args {
				email, publicID, userID := gofakeit.Email(), gofakeit.UUID(), 1
				ctx := helpers.SetUserIDCtx(context.Background(), publicID)
				code := "000000"

				repo.EXPECT().GetUserByPublicID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       int64(userID),
					Email:    email,
					PublicID: publicID,
				}, nil)
				otp.EXPECT().
					GetOTPByCodeAndUserID(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.OTP{
						PublicID:  publicID,
						Code:      code,
						Revoked:   false,
						ExpiresAt: time.Now().Add(5 * time.Minute),
					}, nil)
				otp.EXPECT().RevokeAllOTPsForUser(mock.Anything, mock.Anything, mock.Anything).Return(errMsg)
				repo.EXPECT().SaveRefreshToken(mock.Anything, mock.AnythingOfType("domain.RefreshToken")).Return(nil)

				return args{ctx: ctx, input: &dto.VerifyOTPInput{Purpose: dto.Login, Value: "000000"}}
			},
		},
		{
			name:    "sad case: expired otp",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp) args {
				email, publicID, userID := gofakeit.Email(), gofakeit.UUID(), 1
				ctx := helpers.SetUserIDCtx(context.Background(), publicID)
				code := "000000"

				repo.EXPECT().GetUserByPublicID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       int64(userID),
					Email:    email,
					PublicID: publicID,
				}, nil)
				otp.EXPECT().
					GetOTPByCodeAndUserID(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.OTP{
						PublicID:  publicID,
						Code:      code,
						Revoked:   false,
						ExpiresAt: time.Date(2026, 0o1, 0o1, 11, 30, 45, 98, time.UTC),
					}, nil)

				return args{ctx: ctx, input: &dto.VerifyOTPInput{Purpose: dto.Login, Value: "000000"}}
			},
		},
		{
			name:    "sad case: revoked otp",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp) args {
				email, publicID, userID := gofakeit.Email(), gofakeit.UUID(), 1
				ctx := helpers.SetUserIDCtx(context.Background(), publicID)
				code := "000000"

				repo.EXPECT().GetUserByPublicID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       int64(userID),
					Email:    email,
					PublicID: publicID,
				}, nil)
				otp.EXPECT().
					GetOTPByCodeAndUserID(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(&domain.OTP{
						PublicID:  publicID,
						Code:      code,
						Revoked:   true,
						ExpiresAt: time.Now().Add(5 * time.Minute),
					}, nil)

				return args{ctx: ctx, input: &dto.VerifyOTPInput{Purpose: dto.Login, Value: "000000"}}
			},
		},
		{
			name:    "sad case: get otp by user id failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp) args {
				email, publicID, userID := gofakeit.Email(), gofakeit.UUID(), 1
				ctx := helpers.SetUserIDCtx(context.Background(), publicID)

				repo.EXPECT().GetUserByPublicID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       int64(userID),
					Email:    email,
					PublicID: publicID,
				}, nil)
				otp.EXPECT().
					GetOTPByCodeAndUserID(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errMsg)

				return args{ctx: ctx, input: &dto.VerifyOTPInput{Purpose: dto.Login, Value: "000000"}}
			},
		},
		{
			name:    "sad case: get otp by user id failed - code not found",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, otp *mocks.Mockotp) args {
				email, publicID, userID := gofakeit.Email(), gofakeit.UUID(), 1
				ctx := helpers.SetUserIDCtx(context.Background(), publicID)

				repo.EXPECT().GetUserByPublicID(mock.Anything, mock.Anything).Return(&domain.User{
					ID:       int64(userID),
					Email:    email,
					PublicID: publicID,
				}, nil)
				otp.EXPECT().
					GetOTPByCodeAndUserID(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, repository.ErrNotFound)

				return args{ctx: ctx, input: &dto.VerifyOTPInput{Purpose: dto.Login, Value: "000000"}}
			},
		},
		{
			name:    "sad case: get user by public id failed",
			wantErr: true,
			setup: func(repo *mocks.Mockrepo, _ *mocks.Mockotp) args {
				publicID := gofakeit.UUID()
				ctx := helpers.SetUserIDCtx(context.Background(), publicID)

				repo.EXPECT().GetUserByPublicID(mock.Anything, mock.Anything).Return(nil, errMsg)

				return args{ctx: ctx, input: &dto.VerifyOTPInput{Purpose: dto.Login, Value: "000000"}}
			},
		},
		{
			name:    "sad case: user id not found in context",
			wantErr: true,
			setup: func(_ *mocks.Mockrepo, _ *mocks.Mockotp) args {
				return args{ctx: context.Background(), input: &dto.VerifyOTPInput{Purpose: dto.Login, Value: "000000"}}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, hibp, otp, tasks, cache := mocks.NewMockrepo(
				t,
			), mocks.NewMockhibp(
				t,
			), mocks.NewMockotp(
				t,
			), mocks.NewMockbackgroundTasks(
				t,
			), mocks.NewMockcache(
				t,
			)
			u := New(repo, hibp, otp, tasks, cache)
			args := tt.setup(repo, otp)

			_, gotErr := u.VerifyOTP(args.ctx, args.input)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("VerifyOTP() error = %v: wantErr = %v", gotErr, tt.wantErr)

				return
			}
		})
	}
}
