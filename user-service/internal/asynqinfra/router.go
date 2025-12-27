package asynqinfra

import (
	"github.com/hibiken/asynq"
	"github.com/vanjmali/spotlite/user-service/internal/tasks"
	"github.com/vanjmali/spotlite/user-service/internal/worker"
)

func NewTaskRouter(w *worker.UserWorker) *asynq.ServeMux {
	mux := asynq.NewServeMux()

	mux.HandleFunc(tasks.TypePasswordExpiryCheck, w.HandleExpiryCheck)
	mux.HandleFunc(tasks.TypeSendExpiryEmail, w.HandleSendExpiryEmail)

	return mux
}
