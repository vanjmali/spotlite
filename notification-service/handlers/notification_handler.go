package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/notifications/infrastructure"
	"github.com/vanjmali/spotlite/notifications/services"
)

type NotificationHandler struct {
	s *services.NotificationService
	b *infrastructure.Broker
}

func NewNotificationHandler(s *services.NotificationService, b *infrastructure.Broker) *NotificationHandler {
	h := NotificationHandler{s: s, b: b}
	return &h
}

func (h *NotificationHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	n, err := h.s.CreateNotification(r.Context())

	if err != nil {
		_ = respond.InternalServerError(w)
		return
	}

	userID := middlewares.GetUserIdFromContext(r.Context())
	np, err := json.Marshal(n)
	if err != nil {
		_ = respond.InternalServerError(w)
		return
	}

	h.b.Broadcast <- infrastructure.NewNotification(userID, np)

	respond.NoContent(w)
}

// HandleSubscribe function is used to handle client subscription requests and opens a one way connection
// from server to client.
func (h *NotificationHandler) Subscribe(w http.ResponseWriter, r *http.Request) {

	// removing the write timeout for this request only so the SSE connection
	// can
	rc := http.NewResponseController(w)
	err := rc.SetWriteDeadline(time.Time{})
	if err != nil {
		_ = respond.InternalServerError(w)
	}

	userID := middlewares.GetUserIdFromContext(r.Context())

	if userID == "" {
		_ = respond.Unauthorized(w)
		return
	}

	setSSEHeaders(w)

	// Initializes a new client connection,
	notifChan := make(chan []byte, 10)
	cc := infrastructure.NewClientConnection(userID, notifChan)

	// Adds the client connection event into the ConnectionEvents channel which
	// is used as a queue,
	h.b.ConnectionEvents <- infrastructure.ClientEvent{
		Action: infrastructure.ClientConnect,
		Conn:   cc,
	}

	// schedule the connection closing for the end of the function lifetime
	defer func() {
		h.b.ConnectionEvents <- infrastructure.ClientEvent{
			Action: infrastructure.ClientDisconnect,
			Conn:   cc,
		}
	}()

	// we need to check if the response writer implements the http Flusher interface,
	// normally http responses are formed by waiting to gather as much data as possible
	// and then sending it all at once, but in SSE we need to constantly flush new
	// changes
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// force flush to establish the connection,
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// flush which makes proxies, middlewares acknowledge that the connection is alive and open and from here
	// events can be sent :D
	fmt.Fprintf(w, ":connected\n\n")
	flusher.Flush()

	// ticker will send signals every 5 minutes and will help us ping the client to keep
	// the connection open
	ticker := time.NewTicker(5 * time.Second)

	// schedule ticker stopping for the end of the function lifetime
	defer ticker.Stop()

	// notify will recieve a signal when the context is done and the connection is closed
	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			if _, err := fmt.Fprintf(w, "event: ping\ndata: \n\n"); err != nil {
				// if the ping wasn't successful return which will call all defer calls
				return
			}
			flusher.Flush()
		case msg := <-notifChan:
			// the browser strips away data: %s\n\n
			_, err := fmt.Fprintf(w, "data: %s\n\n", msg)
			if err != nil {
				return
			}
			// if everything goes as planned
			flusher.Flush()
		}
	}
}

func (h *NotificationHandler) GetUserInbox(w http.ResponseWriter, r *http.Request) {
	ns, err := h.s.FindInboxByUserID(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrMissingUserID):
			_ = respond.BadRequest(w)
			return
		default:
			_ = respond.InternalServerError(w)
			return
		}
	}

	_ = respond.OkJson(w, ns)
}

// helpers
func setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
}

type NotificationEvent struct {
	UserID         string     `json:"user_id"`
	CreatedAt      time.Time  `json:"created_at"`
	NotificationID gocql.UUID `json:"notification_id"`
	Message        string     `json:"message"`
}
