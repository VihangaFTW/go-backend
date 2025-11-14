package worker

import (
	"context"
	"encoding/json"
	"fmt"

	db "github.com/VihangaFTW/Go-Backend/db/sqlc"
	"github.com/VihangaFTW/Go-Backend/util"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {

	// serialize payload struct into json (NewTask takes payload as []byte)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	// create redis task
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	// send task to redis queue
	info, err := distributor.client.EnqueueContext(ctx, task)

	if err != nil {
		return fmt.Errorf("failed to enqueue task into redis queue: %w", err)
	}

	// log task details
	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).
		Msg("enqueued task")

	return nil

}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail) error {
	user, err := processor.store.GetUser(ctx, payload.Username)

	if err != nil {
		//? allow retries in case db transactions have not commited before task is processed
		// if err == sql.ErrNoRows {
		// 	return fmt.Errorf("user doesnt exist: %w", err)
		// }
		return fmt.Errorf("failed to get user: %w", err)
	}

	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})

	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	verifyUrl := fmt.Sprintf("http://localhost:8080/v1/verify_email?email_id=%d&secret_code=%s", verifyEmail.ID, verifyEmail.SecretCode)

	subject := "Welcome to SimpleBank"

	content := fmt.Sprintf(`
	Hello %s,<br/>
	Thank you for registering with us!<br/>
	Please <a href="%s">click here</a> to verify your email address.<br/>
	`, user.FullName, verifyUrl)

	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)

	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	log.Info().
		Str("username", payload.Username).
		Str("email", user.Email).
		Msg("processed task")

	return nil
}
