package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/analytics-service/handlers"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
)

// HandleRequests configures and returns the HTTP handler with all routes
func HandleRequests(ah *handlers.AnalyticsHandler) http.Handler {
	r := mux.NewRouter()

	// Health check endpoint
	middlewares.HandleHealthz(r)

	// API routes with authentication
	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "analytics-service")

	// Analytics endpoints (require authentication)
	api.Handle("/analytics/{userID}",
		middlewares.RequireAuthenticated(http.HandlerFunc(ah.HandleGetUserAnalytics))).
		Methods(http.MethodGet)

	api.Handle("/activity-history/{userID}",
		middlewares.RequireAuthenticated(http.HandlerFunc(ah.HandleGetActivityHistory))).
		Methods(http.MethodGet)

	return r
}
