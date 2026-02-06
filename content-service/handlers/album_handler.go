package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/services"
)

// AlbumHandler wires HTTP handlers to the album service and validators.
type AlbumHandler struct {
	s *services.AlbumService
	v *validator.Validate
}

// NewAlbumHandler creates and returns a new AlbumHandler with the provided service and validator.
func NewAlbumHandler(s services.AlbumService, v validator.Validate) *AlbumHandler {
	h := AlbumHandler{s: &s, v: &v}
	return &h
}

// HandleCreateAlbum handles HTTP POST requests to create a new album.
func (h *AlbumHandler) HandleCreateAlbum(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateAlbumDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s invalid request body: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid request body"))
		}
		return
	}

	if err := h.s.Create(r.Context(), &req); err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, respond.ErrorMessage("Invalid ID format"))
			return
		case errors.Is(err, services.ErrSongNotFound):
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrArtistNotFound):
			_ = respond.NotFound(w)
			return
		default:
			log.Printf("trace_id=%s failed to create album: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.InternalServerError(w)
			return
		}
	}

	respond.NoContent(w)
}

// HandleGetAlbumById handles HTTP GET requests to retrieve a single album by its ID.
func (h *AlbumHandler) HandleGetAlbumById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	album, err := h.s.FindAlbumByID(r.Context(), id)
	if err != nil {
		log.Printf("trace_id=%s failed to get album: %v", telemetry.TraceID(r.Context()), err)

		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, respond.ErrorMessage("Invalid album ID format"))
			return
		case errors.Is(err, services.ErrAlbumNotFound):
			_ = respond.NotFound(w)
			return
		default:
			_ = respond.InternalServerError(w)
			return
		}
	}
	if err := respond.OkJson(w, album); err != nil {
		log.Printf("trace_id=%s failed to write get album response: %v", telemetry.TraceID(r.Context()), err)
	}
}

// HandleAddAlbumSongs handles HTTP POST requests to add songs to an album.
func (h *AlbumHandler) HandleAddAlbumSongs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var dto dtos.AddAlbumSongsDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &dto); !ok {
		if err != nil {
			log.Printf("trace_id=%s invalid request body: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid request body"))
		}
		return
	}

	updatedAlbum, err := h.s.AddSongsToAlbum(r.Context(), id, dto)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		log.Printf("trace_id=%s invalid album id: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid album id"))
	case errors.Is(err, services.ErrAlbumNotFound):
		log.Printf("trace_id=%s album not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case errors.Is(err, services.ErrSongNotFound):
		log.Printf("trace_id=%s song not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case err != nil:
		log.Printf("trace_id=%s failed to add album songs: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, updatedAlbum); err != nil {
		log.Printf("trace_id=%s failed to write add album songs response: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}
}

// HandleGetAlbumSongs handles HTTP GET requests to list songs for an album.
func (h *AlbumHandler) HandleGetAlbumSongs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	songs, err := h.s.GetAlbumSongs(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		log.Printf("trace_id=%s invalid album id: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid album id"))
	case errors.Is(err, services.ErrAlbumNotFound):
		log.Printf("trace_id=%s album not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case err != nil:
		log.Printf("trace_id=%s failed to list album songs: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, songs); err != nil {
		log.Printf("trace_id=%s failed to write album songs response: %v", telemetry.TraceID(r.Context()), err)
	}
}

// HandleDeleteAlbumSong handles HTTP DELETE requests to remove a single song from an album.
func (h *AlbumHandler) HandleDeleteAlbumSong(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	songId := vars["songId"]

	err := h.s.RemoveSongFromAlbum(r.Context(), id, songId)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		log.Printf("trace_id=%s invalid album/song id: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid album or song id"))
	case errors.Is(err, services.ErrAlbumNotFound):
		log.Printf("trace_id=%s album not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case errors.Is(err, services.ErrSongNotFound):
		log.Printf("trace_id=%s song not found in album: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case err != nil:
		log.Printf("trace_id=%s failed to delete album song: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}

// HandleUpdateAlbum handles HTTP PATCH requests to update an existing album.
func (h *AlbumHandler) HandleUpdateAlbum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var dto dtos.UpdateAlbumDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &dto); !ok {
		if err != nil {
			log.Printf("trace_id=%s invalid request body: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid request body"))
		}
		return
	}
	updatedAlbum, err := h.s.UpdateAlbum(r.Context(), id, dto)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		log.Printf("trace_id=%s invalid album id: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid album id"))
	case errors.Is(err, services.ErrAlbumNotFound):
		log.Printf("trace_id=%s album not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case errors.Is(err, services.ErrSongNotFound):
		log.Printf("trace_id=%s song not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case errors.Is(err, services.ErrArtistNotFound):
		log.Printf("trace_id=%s artist not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case err != nil:
		log.Printf("trace_id=%s failed to update album: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, updatedAlbum); err != nil {
		log.Printf("trace_id=%s failed to write update album response: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}
}

// HandleDeleteAlbum handles HTTP DELETE requests to delete an album.
func (h *AlbumHandler) HandleDeleteAlbum(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.s.DeleteAlbum(r.Context(), id)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		log.Printf("trace_id=%s invalid album id: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid album id"))
	case errors.Is(err, services.ErrAlbumNotFound):
		log.Printf("trace_id=%s album not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case err != nil:
		log.Printf("trace_id=%s failed to delete album: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}

// HandleGetAlbums handles HTTP GET requests to retrieve a paginated list of albums with optional filtering.
//

func (h *AlbumHandler) HandleGetAlbums(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	p := pagination.ParsePagination(q)
	query := services.AlbumsQuery{
		Page:     p.Page,
		Size:     p.Size,
		Title:    q.Get("title"),
		Genres:   q.Get("genres"),
		GenreID:  q.Get("genre_id"),
		ArtistId: q.Get("artist_id"),
	}

	handleListResponse(w, r, "albums", func(ctx context.Context) (any, error) {
		return h.s.GetAlbums(ctx, query)
	})
}
