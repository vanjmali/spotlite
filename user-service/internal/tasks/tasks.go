package tasks

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypePasswordExpiryCheck = "auth:password_expiry_check"
	TypeSendExpiryEmail     = "auth:send_expiry_email"
)

type SendEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

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
	payload, err := json.Marshal(SendEmailPayload{UserID: id, Email: email})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendExpiryEmail, payload), nil
}
