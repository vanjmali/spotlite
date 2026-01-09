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

type ArtistHandler struct {
	s *services.ArtistService
	v *validator.Validate
}

func NewArtistHandler(s services.ArtistService, v validator.Validate) *ArtistHandler {
	h := ArtistHandler{s: &s, v: &v}
	return &h
}

func (h *ArtistHandler) HandleCreateArtist(w http.ResponseWriter, r *http.Request) {
	var req dtos.ArtistDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s failed to process create artist request: %v", telemetry.TraceID(r.Context()), err)
		}
		return
	}

	// Initializes artist creation after decoding went well
	// I left this like this because maybe in future we will add some validation logic that can be easily added via switch/case
	err := h.s.Create(r.Context(), &req)
	if err != nil {
		var msg string = "An unexpected error has occurred"

		if msg != "" {
			_ = respond.Conflict(w, msg)
			return
		}

		log.Printf("trace_id=%s failed to create artist: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}

func (h *ArtistHandler) HandleGetArtistById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	artist, err := h.s.FindArtistByID(r.Context(), id)
	if err != nil {
		log.Printf("trace_id=%s failed to get artist: %v", telemetry.TraceID(r.Context()), err)

		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, "Invalid artist ID format")
			return
		case errors.Is(err, services.ErrArtistNotFound):
			_ = respond.NotFound(w)
			return
		default:
			_ = respond.InternalServerError(w)
			return
		}
	}

	if err := respond.OkJson(w, artist); err != nil {
		log.Printf("trace_id=%s failed to write get artist response: %v", telemetry.TraceID(r.Context()), err)
	}
}
