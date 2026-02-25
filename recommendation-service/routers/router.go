package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/recommendation-service/handlers"
)

// HandleRequests wires HTTP routes to recommendation handlers.
func HandleRequests(h *handlers.RecommendationHandler) http.Handler {
	r := mux.NewRouter()
	middlewares.HandleHealthz(r)

	// Create a subrouter for API routes to attach telemetry
	// and other middlewares if needed.
	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "recommendation-service")

	// Personalized recommendations endpoint - requires authentication
	api.Handle(
		"/",
		middlewares.RequireAuthenticated(http.HandlerFunc(h.HandleGetRecommendations)),
	).Methods("GET")

	// Trending songs endpoint - public, no authentication required
	api.Handle(
		"/trending",
		http.HandlerFunc(h.HandleGetTrendingSongs),
	).Methods("GET")

	return r
}
