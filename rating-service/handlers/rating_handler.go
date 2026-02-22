package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	commondtos "github.com/vanjmali/spotlite/common-lib/dtos"
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

const LIMIT_DEFAULT = 10

// RatingHandler wires HTTP handlers to the rating service and validators.
type RatingHandler struct {
	s *services.RatingService
	v *validator.Validate
}

// NewRatingHandler creates and returns a new RatingHandler with the provided service and validator.
func NewRatingHandler(s services.RatingService, v validator.Validate) *RatingHandler {
	h := RatingHandler{s: &s, v: &v}

	return &h
}

// HandleCreateRating handles HTTP POST requests to create a new rating.
func (h *RatingHandler) HandleCreateRating(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateRatingDto

	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "failed to process create rating request: %v", err)
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
			_ = respond.Conflict(w, respond.ErrorMessageWithCode("Rating already exists for this song", "rating_exists"))
			return
		default:
			logging.Errorf(r.Context(), "failed to create rating: %v", err)
			_ = respond.InternalServerError(w)
			return
		}
	}
	respond.NoContent(w)
}

// HandleDeleteRating handles HTTP DELETE requests to delete an existing rating by its ID.
func (h *RatingHandler) HandleDeleteRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ratingIDStr := vars["ratingID"]

	ratingID, err := primitive.ObjectIDFromHex(ratingIDStr)
	if err != nil {
		logging.Warnf(r.Context(), "failed to process delete rating request: %v", err)
		_ = respond.BadRequest(w, respond.ErrorMessage("Invalid rating ID"))
		return
	}

	err = h.s.DeleteRating(ratingID, r.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrRatingNotFound):
			logging.Warnf(r.Context(), "failed to process delete rating request: %v", err)
			_ = respond.NotFound(w)
			return
		case errors.Is(err, services.ErrRatingForbidden):
			logging.Warnf(r.Context(), "failed to process delete rating request: %v", err)
			_ = respond.Forbidden(w, respond.ErrorMessage("You can only delete your own rating"))
			return
		default:
			logging.Errorf(r.Context(), "failed to process delete rating request: %v", err)
			_ = respond.InternalServerError(w)
		}
		return
	}

	respond.NoContent(w)
}

// HandleGetRatingsBySongID handles HTTP GET requests to retrieve ratings for a specific song.
func (h *RatingHandler) HandleGetRatingsBySongID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songIDStr := vars["songID"]

	var limit int
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil {
			if l < 1 || l > 200 {
				limit = LIMIT_DEFAULT
			} else {
				limit = l
			}
		}
	}

	cursor := r.URL.Query().Get("cursor")

	ratings, nextCursor, err := h.s.GetRatingBySong(r.Context(), songIDStr, limit, cursor)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidSongID):
			logging.Warnf(r.Context(), "failed to process get ratings request: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid song ID"))
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

// HandleGetRatingsByUserID handles HTTP GET requests to retrieve ratings for a specific user.
func (h *RatingHandler) HandleGetRatingsByUserID(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	p := pagination.ParsePagination(q)
	query := services.RatingsQuery{
		Page:   p.Page,
		Size:   p.Size,
		UserID: mux.Vars(r)["userID"],
	}

	commondtos.HandleListResponse(w, r, "ratings", func(ctx context.Context) (any, error) {
		return h.s.GetRatingByUser(ctx, query)
	})
}

// HandleUpdateRating handles HTTP PATCH requests to update an existing rating.
func (h *RatingHandler) HandleUpdateRating(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["ID"]

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
		logging.Warnf(r.Context(), "failed to process update rating request: %v", err)
		_ = respond.BadRequest(w, respond.ErrorMessage("Invalid rating ID format"))
		return
	case errors.Is(err, services.ErrNoFieldsToUpdate):
		logging.Warnf(r.Context(), "failed to process update rating request: %v", err)
		_ = respond.BadRequest(w)
		return
	case errors.Is(err, services.ErrRatingForbidden):
		logging.Warnf(r.Context(), "failed to process update rating request: %v", err)
		_ = respond.Forbidden(w, respond.ErrorMessage("You can only edit your own rating"))
		return
	case errors.Is(err, services.ErrRatingNotFound):
		logging.Warnf(r.Context(), "failed to process update rating request: %v", err)
		_ = respond.NotFound(w)
		return
	case err != nil:
		logging.Errorf(r.Context(), "failed to process update rating request: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, updatedRating); err != nil {
		logging.Errorf(r.Context(), "failed to write update rating response: %v", err)
	}
}

// HandleGetAverageRatingBySongID handles HTTP GET requests to retrieve the average rating for a specific song.
func (h *RatingHandler) HandleGetAverageRatingBySongID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songIDStr := vars["songID"]

	summary, err := h.s.GetAverageRatingBySongID(r.Context(), songIDStr)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			logging.Warnf(r.Context(), "failed to process get average rating request: %v", err)
			_ = respond.BadRequest(w, respond.ErrorMessage("invalid song ID format"))
			return
		default:
			logging.Errorf(r.Context(), "failed to process get average rating request: %v", err)
			_ = respond.InternalServerError(w)
			return
		}
	}

	_ = respond.OkJson(w, summary)
}
