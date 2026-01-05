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

	r.HandleFunc("/create-artist", h.HandleCreateArtist).Methods("POST")
	return r
}
