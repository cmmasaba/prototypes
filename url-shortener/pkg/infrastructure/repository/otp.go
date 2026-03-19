package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/domain"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

// CreateOTP returns nil on success.
func (r *Repository) CreateOTP(ctx context.Context, input *domain.OTP) error {
	ctx, span := telemetry.Trace(ctx, packageName, "CreateOTP")
	defer span.End()

	err := r.db.CreateOTP(ctx, sqlc.CreateOTPParams{
		UserID:    input.User,
		Code:      input.Code,
		ExpiresAt: timeToTimestamptz(&input.ExpiresAt),
		Valid:     input.Valid,
		Purpose:   input.Purpose.String(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "saving OTP failed", "err", err)
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}

// GetOTPByCodeAndUser returns *[domain.OTP] matching code, user and purpose.
func (r *Repository) GetOTPByCodeAndUser(ctx context.Context, code string, user int64, purpose dto.OTPPurpose) (*domain.OTP, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetOTPByCodeAndUser")
	defer span.End()

	res, err := r.db.GetOTPByCodeAndUserID(ctx, sqlc.GetOTPByCodeAndUserIDParams{
		Code:    code,
		UserID:  user,
		Purpose: purpose.String(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		slog.ErrorContext(ctx, "get otp by code and user failed", "err", err)
		telemetry.RecordError(span, err)

		return nil, err
	}

	return &domain.OTP{
		User:      res.UserID,
		Valid:     res.Valid,
		Code:      res.Code,
		Purpose:   purpose,
		ExpiresAt: *timestamptzToTime(res.ExpiresAt),
	}, nil
}

// RevokeOTP returns nil on success.
func (r *Repository) RevokeOTP(ctx context.Context, code string) error {
	ctx, span := telemetry.Trace(ctx, packageName, "RevokeOTP")
	defer span.End()

	err := r.db.RevokeOTP(ctx, code)
	if err != nil {
		slog.ErrorContext(ctx, "revoke otp failed", "err", err)
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}
