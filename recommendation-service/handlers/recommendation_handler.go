package handlers

import (
	"errors"
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/recommendation-service/services"
)

type RecommendationHandler struct {
	recommendationService *services.RecommendationService
}

func NewRecommendationHandler(rs *services.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{
		recommendationService: rs,
	}
}

func (h *RecommendationHandler) SubscriptionBasedRecommendation(w http.ResponseWriter, r *http.Request) {
	rsr, err := h.recommendationService.SubscriptionBasedRecommendation(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			logging.Errorf(r.Context(), "failed to find subscription based recommendation: %v", err)
			_ = respond.BadRequest(w)
			return
		default:
			logging.Errorf(r.Context(), "failed to find subscription based recommendation: %v", err)
			_ = respond.InternalServerError(w)
			return
		}
	}

	if err := respond.OkJson(w, rsr); err != nil {
		logging.Errorf(r.Context(), "failed to write find subscription based recommendation response: %v", err)
	}
}

func (h *RecommendationHandler) LikeBasedRecommendation(w http.ResponseWriter, r *http.Request) {
	rsr, err := h.recommendationService.LikeBasedRecommendation(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			logging.Errorf(r.Context(), "failed to find like based recommendation: %v", err)
			_ = respond.BadRequest(w)
			return
		default:
			logging.Errorf(r.Context(), "failed to find like based recommendation: %v", err)
			_ = respond.InternalServerError(w)
			return
		}
	}

	if err := respond.OkJson(w, rsr); err != nil {
		logging.Errorf(r.Context(), "failed to write find like based recommendation response: %v", err)
	}
}
