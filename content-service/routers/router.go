package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/handlers"
)

// HandleRequests wires HTTP routes to user handlers.
func HandleRequests(h *handlers.ArtistHandler) http.Handler {

	r := mux.NewRouter()
	middlewares.HandleHealthz(r)

	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "artist-service")

	api.HandleFunc("/artists", h.HandleListArtists).Methods("GET")
	api.HandleFunc("/artists/{id}", h.HandleGetArtistById).Methods("GET")
	api.HandleFunc("/artists", h.HandleCreateArtist).Methods("POST")
	api.HandleFunc("/artists/{id}", h.HandleUpdateArtist).Methods("PATCH")
	api.HandleFunc("/artists/{id}", h.HandleDelete).Methods("DELETE")

	return r
}
