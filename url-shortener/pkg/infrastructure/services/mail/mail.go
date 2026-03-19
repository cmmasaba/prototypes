// Package mail provides functionality for sending emails.
package mail // nolint: revive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	gomail "gopkg.in/mail.v2"
)

const packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/mail"

var (
	errSMTPClientNotSet   = errors.New("smtp client not set")
	errSenderNotSet       = errors.New("sender address not set")
	errSMTPUsernameNotSet = errors.New("smtp username not set")
	errSMTPPasswordNotSet = errors.New("smtp password not set")
	errSMTPPortNotSet     = errors.New("smtp port not set")
	serviceURL            = os.Getenv("SERVICE_BASE_URL")
)

type Mailer interface {
	Send(ctx context.Context, msg any) (bool, error)
}

type MailerImpl struct {
	smtpClient, sender, smtpUsername, smtpPassword string
	port                                           int
}

type emailVerification struct {
	ServiceURL, OTP string
	Year            int
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

// SendEmailVerification returns true and nil after successfully sending an email.
func (m *MailerImpl) SendEmailVerification(ctx context.Context, recipient, otp string) (bool, error) {
	ctx, span := telemetry.Trace(ctx, packageName, "Send")
	defer span.End()

	email := emailVerification{
		ServiceURL: serviceURL,
		OTP:        otp,
		Year:       time.Now().Year(),
	}

	var b bytes.Buffer

	t := template.Must(template.New("emailVerification").Parse(emailVerificationTemplate()))

	err := t.Execute(&b, email)
	if err != nil {
		slog.ErrorContext(ctx, "template execution failed", "err", err)
		telemetry.RecordError(span, err)
	}

	msg := gomail.NewMessage()

	msg.SetAddressHeader("From", m.sender, "URL Shortener")
	msg.SetHeader("To", recipient)
	msg.SetHeader("Subject", "Verify Your Email")
	msg.SetBody("text/html", b.String())

	d := gomail.NewDialer(m.smtpClient, m.port, m.smtpUsername, m.smtpPassword)

	err = d.DialAndSend(msg)
	if err != nil {
		slog.ErrorContext(ctx, "send email failed", "err", err)
		telemetry.RecordError(span, err)

		return false, err
	}

	return true, nil
}
