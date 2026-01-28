package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/content/entities"
	"github.com/vanjmali/spotlite/content/services"
	"golang.org/x/sync/errgroup"
)

// GlobalSearchHandler handles global search requests across multiple content types.
type GlobalSearchHandler struct {
	genreService  *services.GenreService
	songService   *services.SongService
	albumService  *services.AlbumService
	artistService *services.ArtistService
}

func NewGlobalSearchHandler(gs services.GenreService, ss services.SongService, as services.AlbumService, ars services.ArtistService) *GlobalSearchHandler {
	return &GlobalSearchHandler{
		genreService:  &gs,
		songService:   &ss,
		albumService:  &as,
		artistService: &ars,
	}
}

func (h *GlobalSearchHandler) HandleGlobalSearch(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		respond.BadRequest(w, "Search query 'q' is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	artists := []entities.Artist{}
	albums := []entities.Album{}
	songs := []entities.Song{}
	genres := []entities.Genre{}

	g.Go(func() error {
		res, err := h.genreService.GetGenres(ctx, services.GenresQuery{
			Page: 1,
			Size: 5,
			Name: searchTerm,
		})
		if err == nil && res != nil {
			genres = res.Items
		}
		return err
	})

	g.Go(func() error {
		res, err := h.albumService.GetAlbums(ctx, services.AlbumsQuery{
			Page:  1,
			Size:  5,
			Title: searchTerm,
		})
		if err == nil && res != nil {
			albums = res.Items
		}
		return err
	})

	g.Go(func() error {
		res, err := h.songService.GetSongs(ctx, services.SongsQuery{
			Page:  1,
			Size:  5,
			Title: searchTerm,
		})
		if err == nil && res != nil {
			songs = res.Items
		}
		return err
	})

	g.Go(func() error {
		res, err := h.artistService.GetArtists(ctx, services.ArtistsQuery{
			Page: 1,
			Size: 5,
			Name: searchTerm,
		})
		if err == nil && res != nil {
			artists = res.Items
		}
		return err
	})

	if err := g.Wait(); err != nil {
		_ = respond.InternalServerError(w)
	}

	respond.OkJson(w, map[string]any{
		"genres":  genres,
		"albums":  albums,
		"songs":   songs,
		"artists": artists,
	})
}
