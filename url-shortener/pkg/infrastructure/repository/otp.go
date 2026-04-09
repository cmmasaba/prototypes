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

	publicID, err := stringToPgtypeUUID(input.PublicID)
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), "err", err)
		telemetry.RecordError(span, err)

		return err
	}

	err = r.db.CreateOTP(ctx, sqlc.CreateOTPParams{
		Code:         input.Code,
		UserPublicID: publicID,
		ExpiresAt:    timeToTimestamptz(&input.ExpiresAt),
		Revoked:      input.Revoked,
		Purpose:      input.Purpose.String(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "saving OTP failed", "err", err)
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}

// GetOTPByCodeAndUser returns *[domain.OTP] matching code, user and purpose.
func (r *Repository) GetOTPByCodeAndUser(
	ctx context.Context,
	code string,
	user string,
	purpose dto.OTPPurpose,
) (*domain.OTP, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "GetOTPByCodeAndUser")
	defer span.End()

	publicID, err := stringToPgtypeUUID(user)
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), "err", err)
		telemetry.RecordError(span, err)

		return nil, err
	}

	res, err := r.db.GetOTPByCodeAndUserID(ctx, sqlc.GetOTPByCodeAndUserIDParams{
		Code:         code,
		UserPublicID: publicID,
		Purpose:      purpose.String(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "get otp by code and user failed", "err", err)
		telemetry.RecordError(span, err)

		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &domain.OTP{
		PublicID:  pgtypeUUIDToString(res.UserPublicID),
		Revoked:   res.Revoked,
		Code:      res.Code,
		Purpose:   purpose,
		ExpiresAt: *timestamptzToTime(res.ExpiresAt),
	}, nil
}

// RevokeAllOTPsForUser returns nil on success.
func (r *Repository) RevokeAllOTPsForUser(ctx context.Context, user, purpose string) error {
	ctx, span := telemetry.Trace(ctx, packageName, "RevokeOTP")
	defer span.End()

	publicID, err := stringToPgtypeUUID(user)
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), "err", err)
		telemetry.RecordError(span, err)

		return err
	}

	err = r.db.RevokeAllOTPsForUser(ctx, sqlc.RevokeAllOTPsForUserParams{
		UserPublicID: publicID,
		Purpose:      purpose,
	})
	if err != nil {
		slog.ErrorContext(ctx, "revoke otps failed", "err", err)
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}
