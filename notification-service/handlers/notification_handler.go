package handlers

import (
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/respond"
)

type NotificationHandler struct {
}

type UserHandlerConfig struct {
}

func NewNotificationHandler() *NotificationHandler {
	h := NotificationHandler{}
	return &h
}

func (h *NotificationHandler) HandleTest(w http.ResponseWriter, r *http.Request) {
	respond.NoContent(w)

}
