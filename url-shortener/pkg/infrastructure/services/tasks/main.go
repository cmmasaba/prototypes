// Package tasks implements task queueing and scheduling functionalities.
package tasks

import (
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/helpers"
	email "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/mail"
	"github.com/hibiken/asynq"
)

const packageName = "github.com/cmmasaba/prototypes/urlshortener/pkg/infrastructure/services/taskqueue"

type Queue struct {
	client  *asynq.Client
	mailSvc *email.MailerImpl
	server  *asynq.Server
}

func New(mailSvc *email.MailerImpl) (*Queue, error) {
	conn, err := asynq.ParseRedisURI(helpers.MustGetEnvVar("REDIS_URL"))
	if err != nil {
		return nil, err
	}

	queue := &Queue{
		client:  asynq.NewClient(conn),
		mailSvc: mailSvc,
	}

	queue.startTaskServer(conn)

	return queue, nil
}
