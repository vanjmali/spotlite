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

	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "recommendation-service")

	api.Handle("/subscription-based", middlewares.RequireAuthenticated(h.SubscriptionBasedRecommendation)).Methods("GET")
	api.Handle("/like-based", middlewares.RequireAuthenticated(h.LikeBasedRecommendation)).Methods("GET")

	return r
}
