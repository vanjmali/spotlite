package handlers

import (
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
