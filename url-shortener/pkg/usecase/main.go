// Package usecase encapsulates all usecase objects.
package usecase

import (
	"context"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/auth"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/healthcheck"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/usecase/shortener"
)

type Usecase struct {
	health    *healthcheck.UsecaseImpl
	auth      *auth.UsecaseImpl
	shortener *shortener.UsecaseImpl
}

func New(health *healthcheck.UsecaseImpl, auth *auth.UsecaseImpl) *Usecase {
	return &Usecase{
		health: health,
		auth:   auth,
	}
}

func (u *Usecase) HealthCheck(ctx context.Context) map[string]bool {
	return u.health.HealthCheck(ctx)
}

func (u *Usecase) CreateUserEmailPassword(
	ctx context.Context,
	input *dto.EmailPasswordUserInput,
) (*dto.OTPRequiredResponse, error) {
	return u.auth.CreateUserEmailPassword(ctx, input)
}

func (u *Usecase) ValidatePasswordStrength(ctx context.Context, input *dto.ValidatePasswordInput) bool {
	return u.auth.ValidatePasswordStrength(ctx, input)
}

func (u *Usecase) Login(ctx context.Context, input *dto.LoginInput) (*dto.OTPRequiredResponse, error) {
	return u.auth.Login(ctx, input)
}

func (u *Usecase) RefreshAccessToken(
	ctx context.Context,
	refreshToken string,
) (*dto.RefreshAccessTokenResponse, error) {
	return u.auth.RefreshAccessToken(ctx, refreshToken)
}

func (u *Usecase) VerifyOTP(ctx context.Context, input *dto.VerifyOTPInput) (*dto.AuthResponse, error) {
	return u.auth.VerifyOTP(ctx, input)
}

func (u *Usecase) InitOAuthFlow(ctx context.Context, provider dto.OAuthProvider, returnTo string) (string, error) {
	return u.auth.InitOAuthFlow(ctx, provider, returnTo)
}

func (u *Usecase) OAuthFlowCallback(ctx context.Context, origin, code string) (*dto.AuthResponse, string, error) {
	return u.auth.OAuthFlowCallback(ctx, origin, code)
}

func (u *Usecase) Logout(ctx context.Context) error {
	return u.auth.Logout(ctx)
}

func (u *Usecase) RequestNewOTP(ctx context.Context, publicUserID, recipient string, purpose dto.OTPPurpose) error {
	return u.auth.RequestNewOTP(ctx, publicUserID, recipient, purpose)
}

func (u *Usecase) ShortenURL(ctx context.Context, input *dto.ShortenURLInput) (*dto.ShortenURLResponse, error) {
	return u.shortener.ShortenURL(ctx, input)
}

func (u *Usecase) GetOriginalURL(ctx context.Context, code string) (string, error) {
	return u.shortener.GetOriginalURL(ctx, code)
}
