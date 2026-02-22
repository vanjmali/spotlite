package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/rating-service/dtos"
	"github.com/vanjmali/spotlite/rating-service/mappers"
	"github.com/vanjmali/spotlite/rating-service/repositories"
	"github.com/vanjmali/spotlite/rating-service/services"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RatingHandler struct {
	s *services.RatingService
	v *validator.Validate
}

func NewRatingHandler(s services.RatingService, v validator.Validate) *RatingHandler {
	h := RatingHandler{s: &s, v: &v}

	return &h
}

func (h *RatingHandler) HandleCreateRating(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateRatingDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "failed to process subscribe request: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid request body"))
		}
		return
	}

	if err := h.s.CreateRating(&req, r.Context()); err != nil {
		switch {
		case errors.Is(err, mappers.ErrRatingMapping):
			logging.Warnf(r.Context(), "failed to process rating request: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage(err.Error()))
			return
		case errors.Is(err, services.ErrSongNotFound):
			logging.Warnf(r.Context(), "failed to process rating request: %v", err)
			_ = respond.NotFound(w)
			return
		case errors.Is(err, repositories.ErrRatingAlreadyExists):
			logging.Warnf(r.Context(), "failed to process rating request: %v", err)
			_ = respond.Conflict(w, respond.ErrorMessageWithCode("Rating already exists for this song.", "rating_exists"))
			return
		default:
			logging.Errorf(r.Context(), "failed to create rating: %v", err)
			_ = respond.InternalServerError(w)
			return
		}
	}
	respond.NoContent(w)
}

func (h *RatingHandler) HandleDeleteRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ratingIDStr := vars["ratingID"]
	log.Println(ratingIDStr)

	ratingID, err := primitive.ObjectIDFromHex(ratingIDStr)
	if err != nil {
		logging.Warnf(r.Context(), "failed to process delete rating request: %v", err)
		_ = respond.BadRequest(w, respond.ErrorMessage("Invalid rating ID."))
		return
	}

	err = h.s.DeleteRating(ratingID, r.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRatingNotFound):
			logging.Warnf(r.Context(), "failed to process delete rating request: %v", err)
			_ = respond.NotFound(w)
			return
		default:
			logging.Errorf(r.Context(), "failed to process delete rating request: %v", err)
			_ = respond.InternalServerError(w)
		}
		return
	}

	respond.NoContent(w)
}

func (h *RatingHandler) HandleGetRatingsBySongID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songIDStr := vars["song_id"]

	limit := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			if parsed < 1 {
				limit = 1
			} else if parsed > 200 {
				limit = 200
			} else {
				limit = parsed
			}
		}
	}

	cursor := r.URL.Query().Get("cursor")

	ratings, nextCursor, err := h.s.GetRatingBySong(r.Context(), songIDStr, limit, cursor)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidSongID):
			logging.Warnf(r.Context(), "failed to process get ratings request: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid song ID."))
			return
		default:
			logging.Errorf(r.Context(), "failed to process get ratings request: %v", err)
			_ = respond.InternalServerError(w)
			return
		}
	}

	response := map[string]interface{}{
		"items": ratings,
	}

	if nextCursor != "" {
		response["nextCursor"] = nextCursor
	}

	_ = respond.OkJson(w, response)
}

func (h *RatingHandler) HandleGetRatingsByUserID(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	p := pagination.ParsePagination(q)
	query := services.RatingsQuery{
		Page:   p.Page,
		Size:   p.Size,
		UserID: q.Get("user_id"),
	}

	handleListResponse(w, r, "ratings", func(ctx context.Context) (any, error) {
		return h.s.GetRatingByUser(ctx, query)
	})
}

func (h *RatingHandler) HandleUpdateRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var dto dtos.UpdateRatingDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &dto); !ok {
		if err != nil {
			logging.Errorf(r.Context(), "failed to process update rating request: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid request body"))
		}
		return
	}

	updatedRating, err := h.s.UpdateRating(r.Context(), id, dto)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		_ = respond.BadRequest(w, respond.ErrorMessage("Invalid rating ID format"))
		return
	case errors.Is(err, services.ErrRatingNotFound):
		_ = respond.NotFound(w)
		return
	case err != nil:
		logging.Errorf(r.Context(), "failed to update rating: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, updatedRating); err != nil {
		logging.Errorf(r.Context(), "failed to write update rating response: %v", err)
	}
}
