package handlers

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/rating-service/dtos"
	"github.com/vanjmali/spotlite/rating-service/mappers"
	"github.com/vanjmali/spotlite/rating-service/repositories"
	"github.com/vanjmali/spotlite/rating-service/services"
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
