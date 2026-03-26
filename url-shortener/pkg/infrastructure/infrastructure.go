// Package infrastructure facilitates interaction with repository and external services
package infrastructure

import (
	"log/slog"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/hibp"
	email "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/mail"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/otp"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/tasks"
)

type Infrastructure struct {
	DB         *repository.Repository
	Hibp       *hibp.HIBP
	Mail       *email.MailerImpl
	Otp        *otp.Provider
	TasksQueue *tasks.Queue
}

func New() (*Infrastructure, error) {
	database, err := repository.New()
	if err != nil {
		slog.Error("initialize db repository failed", "err", err)

		return nil, err
	}

	hibp, err := hibp.New()
	if err != nil {
		slog.Error("initialize HIBP service failed", "err", err)

		return nil, err
	}

	mail, err := email.New()
	if err != nil {
		slog.Error("initialize mail service failed", "err", err)

		return nil, err
	}

	queue, err := tasks.New(mail)
	if err != nil {
		slog.Error("initialize background tasks queue failed", "err", err)

		return nil, err
	}

	otp := otp.New(database)

	return &Infrastructure{
		DB:         database,
		Hibp:       hibp,
		Mail:       mail,
		Otp:        otp,
		TasksQueue: queue,
	}, nil
}
