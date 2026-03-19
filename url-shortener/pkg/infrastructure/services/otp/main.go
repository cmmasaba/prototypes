// Package otp provides methods for generating OTPs.
package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
)

const (
	packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/otp"
	otpTTL      = 1 * time.Hour
)

type repository interface {
	CreateOTP(ctx context.Context, input *domain.OTP) error
	GetOTPByCodeAndUser(ctx context.Context, code string, user int64, purpose dto.OTPPurpose) (*domain.OTP, error)
	RevokeOTP(ctx context.Context, code string) error
}

// Provider encapsulates an OTP provider.
type Provider struct {
	otpLength int
	repo      repository
}

// New returns an implementation of [OTPProvider].
func New(repo repository) *Provider {
	return &Provider{
		otpLength: 6,
		repo:      repo,
	}
}

// GenerateOTP returns a 6-digit OTP string and nil error on success.
func (p *Provider) GenerateOTP(ctx context.Context, userID int64, purpose dto.OTPPurpose) (string, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GenerateOTP")
	defer span.End()

	var b strings.Builder

	for range p.otpLength {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			slog.ErrorContext(ctx, "generate random int failed", "err", err)
			telemetry.RecordError(span, err)

			return "", err
		}

		fmt.Fprintf(&b, "%d", num)
	}

	err := p.repo.CreateOTP(ctx, &domain.OTP{
		Code:      b.String(),
		ExpiresAt: time.Now().Add(otpTTL),
		Valid:     true,
		User:      userID,
		Purpose:   purpose,
	})
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// GetOTPByCodeAndUserID returns a *[domain.OTP] that matches the code, userID and purpose.
func (p *Provider) GetOTPByCodeAndUserID(ctx context.Context, code string, userID int64, purpose dto.OTPPurpose) (*domain.OTP, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetOTPByCodeAndUserID")
	defer span.End()

	res, err := p.repo.GetOTPByCodeAndUser(ctx, code, userID, purpose)
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	return &domain.OTP{
		User:      res.User,
		Code:      res.Code,
		ExpiresAt: res.ExpiresAt,
		Valid:     res.Valid,
		Purpose:   res.Purpose,
	}, nil
}

// MarkOTPAsUsed returns nil on success.
func (p *Provider) MarkOTPAsUsed(ctx context.Context, code string) error {
	ctx, span := telemetry.Trace(ctx, packageName, "RevokeOTP")
	defer span.End()

	return p.repo.RevokeOTP(ctx, code)
}
