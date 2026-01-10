package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/handlers"
)

func HandleRequests(h *handlers.ArtistHandler) http.Handler {
	r := mux.NewRouter()
	telemetry.AttachMuxTracing(r, "artist-service")

	r.HandleFunc("/artists", h.HandleCreateArtist).Methods("POST")
	r.HandleFunc("/artists/{id}", h.HandleGetArtistById).Methods("GET")
	r.HandleFunc("/artists/{id}", h.HandleUpdateArtist).Methods("PATCH")
	r.HandleFunc("/artists/{id}", h.HandleDelete).Methods("DELETE")
	return r
}
