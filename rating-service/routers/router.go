package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/rating-service/handlers"
)

func HandleRequests(rh *handlers.RatingHandler) http.Handler {
	r := mux.NewRouter()
	middlewares.HandleHealthz(r)

	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "rating-service")

	api.Handle("/", middlewares.RequireAuthenticated(rh.HandleCreateRating)).Methods("POST")
	api.Handle("/{ratingID}", middlewares.RequireAuthenticated(rh.HandleDeleteRating)).Methods("DELETE")
	api.Handle("/song/{songID}", middlewares.RequireAuthenticated(rh.HandleGetRatingsBySongID)).Methods("GET")

	return r
}
