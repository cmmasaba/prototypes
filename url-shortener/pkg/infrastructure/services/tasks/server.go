package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	email "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/mail"
	"github.com/hibiken/asynq"
)

func (q *Queue) processTask(ctx context.Context, t *asynq.Task) error {
	ctx, span := telemetry.Trace(ctx, packageName, "processTask")
	defer span.End()

	var p EmailDeliveryPayload

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	switch p.EmailType {
	case dto.TypeVerificationEmail:
		otp, ok := p.Opts["otpCode"]
		if !ok {
			return fmt.Errorf("missing otp code for email verification task")
		}

		err := q.mailSvc.SendEmail(ctx, email.EmailVerification, email.Opts{
			Recipient: p.Recipient,
			OTP:       otp,
		})
		if err != nil {
			slog.ErrorContext(ctx, "send email verification otp failed", "err", err)

			return err
		}
	case dto.TypeSecurityAlertEmail:
		err := q.mailSvc.SendEmail(ctx, email.SecurityAlert, email.Opts{
			Recipient: p.Recipient,
		})
		if err != nil {
			slog.ErrorContext(ctx, "send security alert email failed", "err", err)

			return err
		}
	case dto.TypePasswordResetEmail:
		otp, ok := p.Opts["otpCode"]
		if !ok {
			return fmt.Errorf("missing otp code for password reset email task")
		}

		err := q.mailSvc.SendEmail(ctx, email.PasswordReset, email.Opts{
			Recipient: p.Recipient,
			OTP:       otp,
		})
		if err != nil {
			slog.ErrorContext(ctx, "send password reset otp failed", "err", err)

			return err
		}
	case dto.TypeLoginEmail:
		otp, ok := p.Opts["otpCode"]
		if !ok {
			return fmt.Errorf("missing otp code for login email task")
		}

		err := q.mailSvc.SendEmail(ctx, email.Login, email.Opts{
			Recipient: p.Recipient,
			OTP:       otp,
		})
		if err != nil {
			slog.ErrorContext(ctx, "send login otp failed", "err", err)

			return err
		}
	default:
		return fmt.Errorf("unknown email task submitted: %s", p.EmailType)
	}

	return nil
}

func (q *Queue) startTaskServer(conn asynq.RedisConnOpt) {
	server := asynq.NewServer(
		conn,
		asynq.Config{
			Concurrency: 20,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()

	// Register email delivery tasks handlers
	mux.HandleFunc(string(dto.TypeLoginEmail), q.processTask)
	mux.HandleFunc(string(dto.TypePasswordResetEmail), q.processTask)
	mux.HandleFunc(string(dto.TypeSecurityAlertEmail), q.processTask)
	mux.HandleFunc(string(dto.TypeVerificationEmail), q.processTask)

	go func() {
		if err := server.Run(mux); err != nil {
			log.Fatalf("failed to run task workers: %v", err)
		}
	}()

	q.server = server
}

// PingQueue returns true if the task queue connection is reachable.
func (q *Queue) PingQueue(ctx context.Context) bool {
	ctx, span := telemetry.Trace(ctx, packageName, "Ping")
	defer span.End()

	err := q.server.Ping()
	if err != nil {
		slog.ErrorContext(ctx, "check background task runners status failed", "err", err)
		telemetry.RecordError(span, err)
	}

	return err == nil
}
