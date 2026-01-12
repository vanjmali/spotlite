package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/handlers"
)

// HandleRequests wires HTTP routes to user handlers.
func HandleRequests(ah *handlers.ArtistHandler, sh *handlers.SongHandler, alh *handlers.AlbumHandler) http.Handler {

	r := mux.NewRouter()
	middlewares.HandleHealthz(r)

	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "artist-service")

	// Artists endpoints
	api.HandleFunc("/artists", ah.HandleListArtists).Methods("GET")
	api.HandleFunc("/artists/{id}", ah.HandleGetArtistById).Methods("GET")
	api.HandleFunc("/artists", ah.HandleCreateArtist).Methods("POST")
	api.HandleFunc("/artists/{id}", ah.HandleUpdateArtist).Methods("PATCH")
	api.HandleFunc("/artists/{id}", ah.HandleDelete).Methods("DELETE")

	// Songs endpoints
	api.HandleFunc("/songs", sh.HandleCreateSong).Methods("POST")

	// Albums endpoints
	api.HandleFunc("/albums", alh.HandleCreateAlbum).Methods("POST")

	return r
}
