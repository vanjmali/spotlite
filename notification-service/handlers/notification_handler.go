package handlers

import (
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/notifications/services"
)

type NotificationHandler struct {
	s *services.NotificationService
}

func NewNotificationHandler(s *services.NotificationService) *NotificationHandler {
	h := NotificationHandler{s: s}
	return &h
}

func (h *NotificationHandler) HandleCreateNotification(w http.ResponseWriter, r *http.Request) {
	err := h.s.CreateNotification(r.Context())

	if err != nil {
		respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}
