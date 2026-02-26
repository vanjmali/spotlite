package handlers

import (
	"github.com/vanjmali/spotlite/recommendation-service/services"
)

type RecommendationHandler struct {
	services services.Services
}

func NewRecommendationHandler(ss services.Services) *RecommendationHandler {
	return &RecommendationHandler{
		services: ss,
	}
}
