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

// ArtistHandler wires HTTP handlers to the artist service and validators.
type ArtistHandler struct {
	s *services.ArtistService
	v *validator.Validate
}

// NewArtistHandler creates and returns a new ArtistHandler with the provided service and validator.
func NewArtistHandler(s services.ArtistService, v validator.Validate) *ArtistHandler {
	h := ArtistHandler{s: &s, v: &v}
	return &h
}

// HandleCreateArtist handles HTTP POST requests to create a new artist.
func (h *ArtistHandler) HandleCreateArtist(w http.ResponseWriter, r *http.Request) {
	var req dtos.ArtistDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s failed to process create artist request: %v", telemetry.TraceID(r.Context()), err)
			_ = respond.BadRequest(w, "invalid request body")
		}
		return
	}

	// Initializes artist creation after decoding went well
	err := h.s.Create(r.Context(), &req)
	if err != nil {
		log.Printf("trace_id=%s failed to create artist: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}
	respond.NoContent(w)
}

// HandleGetArtistById handles HTTP GET requests to retrieve a single artist by its ID.
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

// HandleUpdateArtist handles HTTP PATCH requests to update an existing artist.
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

// HandleDeleteArtist handles HTTP DELETE requests to delete an artist.
func (h *ArtistHandler) HandleDeleteArtist(w http.ResponseWriter, r *http.Request) {
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

// HandleGetArtists handles HTTP GET requests to retrieve a paginated list of artists with optional filtering.
func (h *ArtistHandler) HandleGetArtists(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	page, err := strconv.Atoi(q.Get("page"))
	if err != nil {
		page = 1
	}

	size, err := strconv.Atoi(q.Get("size"))
	if err != nil {
		size = 10
	}

	dto := dtos.ArtistQueryDto{
		Page:  page,
		Size:  size,
		Name:  q.Get("name"),
		Genre: q.Get("genre"),
	}

	resp, err := h.s.GetArtists(r.Context(), dto)
	if err != nil {
		log.Printf("trace_id=%s failed to list artists: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, resp); err != nil {
		log.Printf("trace_id=%s failed to write list artists response: %v", telemetry.TraceID(r.Context()), err)
	}
}
