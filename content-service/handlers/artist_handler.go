package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

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

func (h *ArtistHandler) HandleUpdateArtist(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var dto dtos.UpdateArtistDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		log.Printf("trace_id=%s failed to decode request body: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, "invalid request body")
		return
	}

	updatedArtist, err := h.s.UpdateArtist(r.Context(), id, dto)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		log.Printf("trace_id=%s invalid artist id: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, "invalid artist id")
	case errors.Is(err, services.ErrArtistNotFound):
		log.Printf("trace_id=%s artist not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case err != nil:
		log.Printf("trace_id=%s failed to update artist: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, updatedArtist); err != nil {
		log.Printf("trace_id=%s failed to write update artist response: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}
}

func (h *ArtistHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	err := h.s.DeleteArtist(r.Context(), id)

	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		log.Printf("trace_id=%s invalid artist id: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.BadRequest(w, "invalid artist id")
	case errors.Is(err, services.ErrArtistNotFound):
		log.Printf("trace_id=%s artist not found: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.NotFound(w)
	case err != nil:
		log.Printf("trace_id=%s failed to delete artist: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}

func (h *ArtistHandler) HandleListArtists(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))

	dto := dtos.ArtistQueryDto{
		Page:  page,
		Size:  size,
		Name:  q.Get("name"),
		Genre: q.Get("genre"),
	}

	resp, err := h.s.GetArtists(r.Context(), dto)
	if err != nil {
		log.Printf(
			"trace_id=%s failed to list artists: %v",
			telemetry.TraceID(r.Context()),
			err,
		)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, resp); err != nil {
		log.Printf(
			"trace_id=%s failed to write list artists response: %v",
			telemetry.TraceID(r.Context()),
			err,
		)
	}
}
