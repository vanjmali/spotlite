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

	// SSE subscribe endpoint
	//
	// it is handled by the base router and avoids being wrapped by the Otel middleware
	// because the Otel middleware doesn't implement flush method which b
	r.Handle("/stream", middlewares.RequireAuthenticated(h.Subscribe)).Methods("GET")

	// Create a subrouter for API routes to attach telemetry
	// and other middlewares if needed.
	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "notification-service")

	api.Handle("/", middlewares.RequireAuthenticated(h.GetUserInbox)).Methods("GET")
	return r
}
