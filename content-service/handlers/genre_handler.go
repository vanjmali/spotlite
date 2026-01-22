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

type GenreHandler struct {
	s *services.GenreService
	v *validator.Validate
}

// NewGenreHandler creates and returns a new GenreHandler with the provided service and validator.
func NewGenreHandler(s services.GenreService, v validator.Validate) *GenreHandler {
	h := GenreHandler{s: &s, v: &v}
	return &h
}

func (h *GenreHandler) HandleCreateGenre(w http.ResponseWriter, r *http.Request) {
	var req dtos.GenreDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s failed to process create genre request: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, "invalid request body")
		}
		return
	}

	if err := h.s.Create(r.Context(), &req); err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "Invalid ID format")
			return
		default:
			log.Printf("trace_id=%s failed to create genre: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.InternalServerError(w)
			return
		}
	}
	respond.NoContent(w)
}

func (h *GenreHandler) HandleGetGenreById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	genre, err := h.s.FindGenreByID(r.Context(), id)
	if err != nil {
		log.Printf("trace_id=%s failed to get genre: %v", telemetry.TraceID(r.Context()), err)

		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "Invalid genre ID format")
			return
		case errors.Is(err, services.ErrGenreNotFound):
			_ = respond.NotFound(w)
			return
		default:
			_ = respond.InternalServerError(w)
			return
		}
	}

	if err := respond.OkJson(w, genre); err != nil {
		log.Printf("trace_id=%s failed to write get genre response: %v", telemetry.TraceID(r.Context()), err)
	}
}
