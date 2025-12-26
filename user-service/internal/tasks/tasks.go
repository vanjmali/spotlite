package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypePasswordExpiryCheck = "auth:password_expiry_check"
	TypeSendExpiryEmail     = "auth:send_expiry_email"
)

type SendEmailPayload struct {
	UserID string
	Email  string
}

func NewPasswordExpiryCheckTask() *asynq.Task {
	return asynq.NewTask(TypePasswordExpiryCheck, nil)
}

func NewSendExpiryEmailTask(id string, email string) (*asynq.Task, error) {
	payload, err := json.Marshal(SendEmailPayload{UserID: id, Email: email})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSendExpiryEmail, payload), nil
}
