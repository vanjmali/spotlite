package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
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
		log.Printf("trace_id=%s search query 'q' is required", telemetry.TraceID(r.Context()))
		_ = respond.BadRequest(w, "Search query 'q' is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	g := new(errgroup.Group)

	genres := []entities.Genre{}
	albums := []entities.Album{}
	songs := []entities.Song{}
	artists := []entities.Artist{}

	g.Go(func() error {
		res, err := h.genreService.GetGenres(ctx, services.GenresQuery{
			Page: 1,
			Size: 5,
			Name: searchTerm,
		})
		if err != nil {
			log.Printf("trace_id=%s genre search failed: %v", telemetry.TraceID(ctx), err)
			return nil
		}
		if res != nil {
			genres = res.Items
		}
		return nil
	})

	g.Go(func() error {
		res, err := h.albumService.GetAlbums(ctx, services.AlbumsQuery{
			Page:  1,
			Size:  5,
			Title: searchTerm,
		})
		if err != nil {
			log.Printf("trace_id=%s album search failed: %v", telemetry.TraceID(ctx), err)
			return nil
		}
		if res != nil {
			albums = res.Items
		}
		return nil
	})

	g.Go(func() error {
		res, err := h.songService.GetSongs(ctx, services.SongsQuery{
			Page:  1,
			Size:  5,
			Title: searchTerm,
		})
		if err != nil {
			log.Printf("trace_id=%s song search failed: %v", telemetry.TraceID(ctx), err)
			return nil
		}
		if res != nil {
			songs = res.Items
		}
		return nil
	})

	g.Go(func() error {
		res, err := h.artistService.GetArtists(ctx, services.ArtistsQuery{
			Page: 1,
			Size: 5,
			Name: searchTerm,
		})
		if err != nil {
			log.Printf("trace_id=%s artist search failed: %v", telemetry.TraceID(ctx), err)
			return nil
		}
		if res != nil {
			artists = res.Items
		}
		return nil
	})

	_ = g.Wait()

	_ = respond.OkJson(w, map[string]any{
		"genres":  genres,
		"albums":  albums,
		"songs":   songs,
		"artists": artists,
	})
}
