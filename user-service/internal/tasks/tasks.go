package tasks

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
	"github.com/vanjmali/spotlite/user-service/internal/payload"
)

const (
	TypePasswordExpiryCheck = "user:password_expiry_check"
	TypeSendExpiryEmail     = "user:send_expiry_email"
)

// NewPasswordExpiryCheckTask task.
func NewPasswordExpiryCheckTask() *asynq.Task {
	return asynq.NewTask(
		TypePasswordExpiryCheck,
		nil,
		asynq.Unique(23*time.Hour),
	)
}

// NewSendExpiryEmailTask task.
func NewSendExpiryEmailTask(id string, email string) (*asynq.Task, error) {
	payload, err := json.Marshal(payload.SendExpiryEmailPayload{UserID: id, Email: email})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendExpiryEmail, payload), nil
}
