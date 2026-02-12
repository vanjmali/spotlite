package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/notification-service/handlers"
)

func HandleRequests(h *handlers.NotificationHandler) http.Handler {
	r := mux.NewRouter()
	middlewares.HandleHealthz(r)

	// Create a subrouter for API routes to attach telemetry
	// and other middlewares if needed.
	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "notification-service")

	api.Handle("/", middlewares.RequireAuthenticated(h.GetUserInbox)).Methods("GET")

	// SSE subscribe endpoint
	api.Handle("/stream", middlewares.RequireAuthenticated(h.Subscribe)).Methods("GET")
	return r
}
