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
	api.HandleFunc("/artists", ah.HandleCreateArtist).Methods("POST")
	api.HandleFunc("/artists/{id}", ah.HandleUpdateArtist).Methods("PATCH")
	api.HandleFunc("/artists/{id}", ah.HandleDeleteArtist).Methods("DELETE")

	// Songs endpoints
	api.HandleFunc("/songs", sh.HandleGetSongs).Methods("GET")
	api.HandleFunc("/songs/{id}", sh.HandleGetSongById).Methods("GET")
	api.HandleFunc("/songs", sh.HandleCreateSongWithAudio).Methods("POST")
	api.HandleFunc("/songs/{id}", sh.HandleUpdateSong).Methods("PATCH")
	api.HandleFunc("/songs/{id}", sh.HandleDeleteSong).Methods("DELETE")
	api.HandleFunc("/songs/{id}/audio", sh.HandleStreamSongAudio).Methods("GET")
	api.HandleFunc("/songs/{id}/audio", sh.HandleUploadSongAudio).Methods("PUT")

	// Albums endpoints
	api.HandleFunc("/albums", alh.HandleGetAlbums).Methods("GET")
	api.HandleFunc("/albums/{id}", alh.HandleGetAlbumById).Methods("GET")
	api.HandleFunc("/albums", alh.HandleCreateAlbum).Methods("POST")
	api.HandleFunc("/albums/{id}", alh.HandleUpdateAlbum).Methods("PATCH")
	api.HandleFunc("/albums/{id}", alh.HandleDeleteAlbum).Methods("DELETE")
	api.HandleFunc("/albums/{id}/songs", alh.HandleAddAlbumSongs).Methods("POST")
	api.HandleFunc("/albums/{id}/songs", alh.HandleGetAlbumSongs).Methods("GET")
	api.HandleFunc("/albums/{id}/songs/{songId}", alh.HandleDeleteAlbumSong).Methods("DELETE")

	// Genres endpoints
	api.HandleFunc("/genres", gh.HandleGetGenres).Methods("GET")
	api.HandleFunc("/genres", gh.HandleCreateGenre).Methods("POST")
	api.HandleFunc("/genres/{id}", gh.HandleGetGenreById).Methods("GET")
	api.HandleFunc("/genres/{id}", gh.HandleUpdateGenre).Methods("PATCH")
	api.HandleFunc("/genres/{id}", gh.HandleDeleteGenre).Methods("DELETE")

	// Global search endpoint
	api.HandleFunc("/search", gsh.HandleGlobalSearch).Methods("GET")

	return r
}
