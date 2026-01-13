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

type SongHandler struct {
	s *services.SongService
	v *validator.Validate
}

func NewSongHandler(s services.SongService, v validator.Validate) *SongHandler {
	h := SongHandler{s: &s, v: &v}
	return &h
}

func (h *SongHandler) HandleCreateSong(w http.ResponseWriter, r *http.Request) {
	var req dtos.SongDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s invalid request body: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, "invalid request body")
		}
		return
	}

	if err := h.s.Create(r.Context(), &req); err != nil {
		log.Printf("trace_id=%s failed to create song: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}

func (h *SongHandler) HandleGetSongById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	song, err := h.s.FindSongById(r.Context(), id)
	if err != nil {
		log.Printf("trace_id=%s failed to get song: %v", telemetry.TraceID(r.Context()), err)

		switch {
		case errors.Is(err, services.ErrSongNotFound):
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "Invalid artist ID format")
			return
		default:
			_ = respond.InternalServerError(w)
			return
		}
	}

	if err := respond.OkJson(w, song); err != nil {
		log.Printf("trace_id=%s failed to write get song response: %v", telemetry.TraceID(r.Context()), err)
	}
}
