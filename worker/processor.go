package worker

import (
	"context"
	"encoding/json"
	"fmt"

	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/mail"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	mailer mail.EmailSender
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer mail.EmailSender) TaskProcessor {

	logger := NewLogger()
	redis.SetLogger(logger)

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error().
					Err(err).
					Str("type", task.Type()).
					Bytes("payload", task.Payload()).
					Msg("process task failed")
			}),
			Logger: logger,
		},
	)

	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	// Wrap the email verify handler with asynq unmarshaling
	mux.HandleFunc(TaskSendVerifyEmail, func(ctx context.Context, task *asynq.Task) error {
		var payload PayloadSendVerifyEmail

		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", fmt.Errorf("%w: %v", asynq.SkipRetry, err))
		}
		return processor.ProcessTaskSendVerifyEmail(ctx, &payload)
	})

	return processor.server.Start(mux)
}
