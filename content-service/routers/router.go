package routers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/handlers"
)

// HandleRequests wires HTTP routes to user handlers.
func HandleRequests(
	ah *handlers.ArtistHandler,
	sh *handlers.SongHandler,
	alh *handlers.AlbumHandler,
	gh *handlers.GenreHandler,
	gsh *handlers.GlobalSearchHandler,
) http.Handler {
	r := mux.NewRouter()
	middlewares.HandleHealthz(r)

	// Create a subrouter for API routes to attach telemetry
	// and other middlewares if needed.
	api := r.PathPrefix("/").Subrouter()
	telemetry.AttachMuxTracing(api, "content-service")

	// Artists endpoints
	api.HandleFunc("/artists", ah.HandleGetArtists).Methods("GET")
	api.HandleFunc("/artists/{id}", ah.HandleGetArtistById).Methods("GET")
	api.Handle("/artists", middlewares.RequireAdmin(ah.HandleCreateArtist)).Methods("POST")
	api.Handle("/artists/{id}", middlewares.RequireAdmin(ah.HandleUpdateArtist)).Methods("PATCH")
	api.Handle("/artists/{id}", middlewares.RequireAdmin(ah.HandleDeleteArtist)).Methods("DELETE")

	// Songs endpoints
	api.HandleFunc("/songs", sh.HandleGetSongs).Methods("GET")
	api.HandleFunc("/songs/{id}", sh.HandleGetSongById).Methods("GET")
	api.Handle("/songs", middlewares.RequireAdmin(sh.HandleCreateSong)).Methods("POST")
	api.Handle("/songs/{id}", middlewares.RequireAdmin(sh.HandleUpdateSong)).Methods("PATCH")
	api.Handle("/songs/{id}", middlewares.RequireAdmin(sh.HandleDeleteSong)).Methods("DELETE")

	// Albums endpoints
	api.HandleFunc("/albums", alh.HandleGetAlbums).Methods("GET")
	api.HandleFunc("/albums/{id}", alh.HandleGetAlbumById).Methods("GET")
	api.Handle("/albums", middlewares.RequireAdmin(alh.HandleCreateAlbum)).Methods("POST")
	api.Handle("/albums/{id}", middlewares.RequireAdmin(alh.HandleUpdateAlbum)).Methods("PATCH")
	api.Handle("/albums/{id}", middlewares.RequireAdmin(alh.HandleDeleteAlbum)).Methods("DELETE")
	api.Handle("/albums/{id}/songs", middlewares.RequireAdmin(alh.HandleAddAlbumSongs)).Methods("POST")
	api.HandleFunc("/albums/{id}/songs", alh.HandleGetAlbumSongs).Methods("GET")
	api.Handle("/albums/{id}/songs/{songId}", middlewares.RequireAdmin(alh.HandleDeleteAlbumSong)).Methods("DELETE")

	// Genres endpoints
	api.HandleFunc("/genres", gh.HandleGetGenres).Methods("GET")
	api.Handle("/genres", middlewares.RequireAdmin(gh.HandleCreateGenre)).Methods("POST")
	api.HandleFunc("/genres/{id}", gh.HandleGetGenreById).Methods("GET")
	api.Handle("/genres/{id}", middlewares.RequireAdmin(gh.HandleUpdateGenre)).Methods("PATCH")
	api.Handle("/genres/{id}", middlewares.RequireAdmin(gh.HandleDeleteGenre)).Methods("DELETE")

	// Global search endpoint
	api.HandleFunc("/search", gsh.HandleGlobalSearch).Methods("GET")

	return r
}
