// Package infrastructure facilitates interaction with repository and external services
package infrastructure

import (
	"fmt"
	"log/slog"

	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/cache"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/repository"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/hibp"
	email "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/mail"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/otp"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/tasks"
	"github.com/redis/go-redis/v9"
)

type Infrastructure struct {
	Cache      *cache.Impl
	DB         *repository.Repository
	Hibp       *hibp.HIBP
	Mail       *email.MailerImpl
	Otp        *otp.Provider
	TasksQueue *tasks.Queue
}

func New(redisClient *redis.Client) (*Infrastructure, error) {
	user := helpers.MustGetEnvVar("POSTGRES_USER")
	password := helpers.MustGetEnvVar("POSTGRES_PASSWORD")
	host := helpers.MustGetEnvVar("POSTGRES_HOST")
	port := helpers.MustGetEnvVar("POSTGRES_PORT")
	sslmode := helpers.MustGetEnvVar("POSTGRES_SSLMODE")
	db := helpers.MustGetEnvVar("POSTGRES_DB")
	connString := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		user,
		password,
		host,
		port,
		db,
		sslmode,
	)

	database, err := repository.New(connString)
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

	cache := cache.New(redisClient)

	otp := otp.New(database)

	return &Infrastructure{
		Cache:      cache,
		DB:         database,
		Hibp:       hibp,
		Mail:       mail,
		Otp:        otp,
		TasksQueue: queue,
	}, nil
}
