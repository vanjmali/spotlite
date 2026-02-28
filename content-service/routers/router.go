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
	api.Handle("/artists", middlewares.RequireAuthenticated(ah.HandleGetArtists)).Methods("GET")
	api.Handle("/artists/{id}", middlewares.RequireAuthenticated(ah.HandleGetArtistById)).Methods("GET")
	api.Handle("/artists", middlewares.RequireAdmin(ah.HandleCreateArtist)).Methods("POST")
	api.Handle("/artists/{id}", middlewares.RequireAdmin(ah.HandleUpdateArtist)).Methods("PATCH")
	api.Handle("/artists/{id}", middlewares.RequireAdmin(ah.HandleDeleteArtist)).Methods("DELETE")

	// Songs endpoints
	api.Handle("/songs", middlewares.RequireAuthenticated(sh.HandleGetSongs)).Methods("GET")
	api.Handle("/songs/{id}", middlewares.RequireAuthenticated(sh.HandleGetSongById)).Methods("GET")
	api.Handle("/songs", middlewares.RequireAdmin(sh.HandleCreateSongWithAudio)).Methods("POST")
	api.Handle("/songs/{id}", middlewares.RequireAdmin(sh.HandleUpdateSong)).Methods("PATCH")
	api.Handle("/songs/{id}", middlewares.RequireAdmin(sh.HandleDeleteSong)).Methods("DELETE")
	api.Handle("/songs/{id}/audio", middlewares.RequireAdmin(sh.HandleUploadSongAudio)).Methods("POST")
	api.Handle("/songs/{id}/audio/signed-url", middlewares.RequireAuthenticated(sh.HandleGetSongAudioSignedURL)).Methods("GET")
	api.HandleFunc("/songs/{id}/audio", sh.HandleStreamSongAudio).Methods("GET")

	// Albums endpoints
	api.Handle("/albums", middlewares.RequireAuthenticated(alh.HandleGetAlbums)).Methods("GET")
	api.Handle("/albums/{id}", middlewares.RequireAuthenticated(alh.HandleGetAlbumById)).Methods("GET")
	api.Handle("/albums", middlewares.RequireAdmin(alh.HandleCreateAlbum)).Methods("POST")
	api.Handle("/albums/{id}", middlewares.RequireAdmin(alh.HandleUpdateAlbum)).Methods("PATCH")
	api.Handle("/albums/{id}", middlewares.RequireAdmin(alh.HandleDeleteAlbum)).Methods("DELETE")
	api.Handle("/albums/{id}/songs", middlewares.RequireAdmin(alh.HandleAddAlbumSongs)).Methods("POST")
	api.Handle("/albums/{id}/songs", middlewares.RequireAuthenticated(alh.HandleGetAlbumSongs)).Methods("GET")
	api.Handle("/albums/{id}/songs/{songId}", middlewares.RequireAdmin(alh.HandleDeleteAlbumSong)).Methods("DELETE")

	// Genres endpoints
	api.Handle("/genres", middlewares.RequireAuthenticated(gh.HandleGetGenres)).Methods("GET")
	api.Handle("/genres", middlewares.RequireAdmin(gh.HandleCreateGenre)).Methods("POST")
	api.Handle("/genres/{id}", middlewares.RequireAuthenticated(gh.HandleGetGenreById)).Methods("GET")
	api.Handle("/genres/{id}", middlewares.RequireAdmin(gh.HandleUpdateGenre)).Methods("PATCH")
	api.Handle("/genres/{id}", middlewares.RequireAdmin(gh.HandleDeleteGenre)).Methods("DELETE")

	// Global search endpoint
	api.Handle("/search", middlewares.RequireAuthenticated(gsh.HandleGlobalSearch)).Methods("GET")

	return r
}
