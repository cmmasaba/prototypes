package tasks

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/cmmasaba/prototypes/telemetry"
	"github.com/cmmasaba/prototypes/urlshortener/pkg/application/dto"
	"github.com/hibiken/asynq"
)

type EmailDeliveryPayload struct {
	EmailType dto.EmailDeliveryType
	Recipient string
	Opts      map[string]string
}

type Priority string

const (
	Critical Priority = "critical"
	Default  Priority = "default"
	Low      Priority = "low"
)

func (t Priority) String() string {
	switch t {
	case Critical:
		return "critical"
	case Default:
		return "default"
	default:
		return "low"
	}
}

// NewEmailDeliveryTask enqueues a new task to send an email and retuens nil on success.
func (q *Queue) NewEmailDeliveryTask(ctx context.Context, input EmailDeliveryPayload, priority Priority) error {
	ctx, span := telemetry.Trace(ctx, packageName, "NewEmailDeliveryTask")
	defer span.End()

	payload, err := json.Marshal(input)
	if err != nil {
		return err
	}

	task := asynq.NewTask(string(input.EmailType), payload)

	info, err := q.client.Enqueue(
		task,
		asynq.MaxRetry(5),
		asynq.Timeout(2*time.Minute),
		asynq.Queue(priority.String()),
	)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "email task enqueued", "id", info.ID, "queue", info.Queue, "type", info.Type)

	return nil
}
