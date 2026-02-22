package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/rating-service/handlers"
)

// HandleRequests wires HTTP routes to user handlers.
func HandleRequests(rh *handlers.RatingHandler) http.Handler {
	r := mux.NewRouter()
	middlewares.HandleHealthz(r)

	// Create a subrouter for API routes to attach telemetry
	// and other middlewares if needed.
	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "rating-service")

	api.Handle("/songs/{songID}", middlewares.RequireAuthenticated(rh.HandleGetRatingsBySongID)).Methods("GET")
	api.Handle("/users/{userID}", middlewares.RequireAuthenticated(rh.HandleGetRatingsByUserID)).Methods("GET")
	api.Handle("/songs/{songID}/average", middlewares.RequireAuthenticated(rh.HandleGetAverageRatingBySongID)).Methods("GET")

	api.Handle("/{id}", middlewares.RequireAuthenticated(rh.HandleUpdateRating)).Methods("PATCH")

	api.Handle("/", middlewares.RequireAuthenticated(rh.HandleCreateRating)).Methods("POST")

	api.Handle("/{ratingID}", middlewares.RequireAuthenticated(rh.HandleDeleteRating)).Methods("DELETE")

	return r
}
