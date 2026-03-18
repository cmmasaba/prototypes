// Package mail provides functionality for sending emails.
package mail // nolint: revive

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	gomail "gopkg.in/mail.v2"
)

const packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/mail"

var (
	errSMTPClientNotSet   = errors.New("smtp client not set")
	errSenderNotSet       = errors.New("sender address not set")
	errSMTPUsernameNotSet = errors.New("smtp username not set")
	errSMTPPasswordNotSet = errors.New("smtp password not set")
	errSMTPPortNotSet     = errors.New("smtp port not set")
)

type Mailer interface {
	Send(ctx context.Context, msg any) (bool, error)
}

type MailerImpl struct {
	smtpClient, sender, smtpUsername, smtpPassword string
	port                                           int
}

func New() (*MailerImpl, error) {
	smtp, ok := os.LookupEnv("SMTP_CLIENT")
	if !ok {
		return nil, errSMTPClientNotSet
	}

	sender, ok := os.LookupEnv("SENDER_EMAIL_ADDRESS")
	if !ok {
		return nil, errSenderNotSet
	}

	username, ok := os.LookupEnv("SMTP_USERNAME")
	if !ok {
		return nil, errSMTPUsernameNotSet
	}

	password, ok := os.LookupEnv("SMTP_PASSWORD")
	if !ok {
		return nil, errSMTPPasswordNotSet
	}

	port, ok := os.LookupEnv("SMTP_PORT")
	if !ok {
		return nil, errSMTPPortNotSet
	}

	smtpPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("convert port to int failed: %w", err)
	}

	return &MailerImpl{
		smtpClient:   smtp,
		sender:       sender,
		smtpUsername: username,
		smtpPassword: password,
		port:         smtpPort,
	}, nil
}

// Send returns true and nil after successfully sending an email.
//
// input should contain: sender addr, subject, body, attachments if any
func (m *MailerImpl) Send(ctx context.Context, input dto.SendMailInput) (bool, error) {
	_, span := telemetry.Trace(ctx, packageName, "Send")
	defer span.End()

	msg := gomail.NewMessage()

	msg.SetHeader("From", m.sender)
	msg.SetHeader("To", input.Address)
	msg.SetHeader("Subject", input.Subject)
	msg.SetBody("text/html", input.Body)

	d := gomail.NewDialer(m.smtpClient, m.port, m.smtpUsername, m.smtpPassword)

	err := d.DialAndSend(msg)
	if err != nil {
		slog.ErrorContext(ctx, "send email failed", "err", err)
		telemetry.RecordError(span, err)

		return false, err
	}

	return true, nil
}
