package handlers

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
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

		log.Printf("trace_id=%s failed to create user: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}
