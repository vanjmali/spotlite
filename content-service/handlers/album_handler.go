package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/dtos"
	"github.com/vanjmali/spotlite/content/services"
)

type AlbumHandler struct {
	s *services.AlbumService
	v *validator.Validate
}

func NewAlbumHandler(s services.AlbumService, v validator.Validate) *AlbumHandler {
	h := AlbumHandler{s: &s, v: &v}
	return &h
}

func (h *AlbumHandler) HandleCreateAlbum(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateAlbumDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s invalid request body: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, "invalid request body")
		}
		return
	}

	if err := h.s.Create(r.Context(), &req); err != nil {

		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "Invalid ID format")
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

func (h *AlbumHandler) HandleGetAlbumById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	album, err := h.s.FindAlbumByID(r.Context(), id)
	if err != nil {
		log.Printf("trace_id=%s failed to get album: %v", telemetry.TraceID(r.Context()), err)

		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "Invalid album ID format")
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
