// Package email provides functionality for sending emails.
package email

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	gomail "gopkg.in/mail.v2"
)

type Type string

const (
	packageName            = "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/mail"
	PasswordReset     Type = "password-reset"
	EmailVerification Type = "email-verification"
	Login             Type = "login"
	SecurityAlert     Type = "security-alert"
)

var emailSubject = map[Type]string{
	PasswordReset:     "Password Reset OTP",
	Login:             "Welcome back",
	EmailVerification: "Verify Your Email",
	SecurityAlert:     "Security Alert",
}

var (
	errOTPNotProvided = errors.New("otp code must be provided")

	serviceURL = os.Getenv("SERVICE_BASE_URL")
)

type MailerImpl struct {
	smtpClient, sender, smtpUsername, smtpPassword                                  string
	port                                                                            int
	loginOTPTmpl, passwordResetOTPTmpl, emailVerificationOTPTmpl, securityAlertTmpl *template.Template
}

type Opts struct {
	Recipient  string
	OTP        string
	ServiceURL string // optional
	Year       int    // optional
}

func New() (*MailerImpl, error) {
	smtp := helpers.MustGetEnvVar("SMTP_CLIENT")
	sender := helpers.MustGetEnvVar("SENDER_EMAIL_ADDRESS")
	username := helpers.MustGetEnvVar("SMTP_USERNAME")
	password := helpers.MustGetEnvVar("SMTP_PASSWORD")
	port := helpers.MustGetEnvVar("SMTP_PORT")

	smtpPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("convert port to int failed: %w", err)
	}

	return &MailerImpl{
		smtpClient:               smtp,
		sender:                   sender,
		smtpUsername:             username,
		smtpPassword:             password,
		port:                     smtpPort,
		passwordResetOTPTmpl:     template.Must(template.New("password-reset").Parse(passwordResetTemplate())),
		loginOTPTmpl:             template.Must(template.New("login").Parse(loginTemplate())),
		emailVerificationOTPTmpl: template.Must(template.New("email-verification").Parse(emailVerificationTemplate())),
		securityAlertTmpl:        template.Must(template.New("security-alert").Parse(securityAlertTemplate())),
	}, nil
}

// SendEmail returns nil on success.
func (m *MailerImpl) SendEmail(ctx context.Context, emailType Type, opts Opts) error {
	ctx, span := telemetry.Trace(ctx, packageName, "SendEmail")
	defer span.End()

	if opts.Recipient == "" {
		return fmt.Errorf("recipient must be provided")
	}

	if opts.ServiceURL == "" {
		opts.ServiceURL = serviceURL
	}

	if opts.Year == 0 {
		opts.Year = time.Now().Year()
	}

	var tmpl *template.Template

	switch emailType {
	case EmailVerification:
		if opts.OTP == "" {
			return errOTPNotProvided
		}

		tmpl = m.emailVerificationOTPTmpl
	case Login:
		if opts.OTP == "" {
			return errOTPNotProvided
		}

		tmpl = m.loginOTPTmpl
	case PasswordReset:
		if opts.OTP == "" {
			return errOTPNotProvided
		}

		tmpl = m.passwordResetOTPTmpl
	case SecurityAlert:
		tmpl = m.securityAlertTmpl
	default:
		return fmt.Errorf("unknown email type provided: %s", emailType)
	}

	var b bytes.Buffer

	err := tmpl.Execute(&b, opts)
	if err != nil {
		slog.ErrorContext(ctx, "template execution failed", "err", err)
		telemetry.RecordError(span, err)

		return err
	}

	msg := gomail.NewMessage()

	msg.SetAddressHeader("From", m.sender, "URL Shortener")
	msg.SetHeader("To", opts.Recipient)
	msg.SetHeader("Subject", emailSubject[emailType])
	msg.SetBody("text/html", b.String())

	d := gomail.NewDialer(m.smtpClient, m.port, m.smtpUsername, m.smtpPassword)

	err = d.DialAndSend(msg)
	if err != nil {
		slog.ErrorContext(ctx, "send email failed", "err", err)
		telemetry.RecordError(span, err)

		return err
	}

	return nil
}
