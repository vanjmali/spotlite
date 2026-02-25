package handlers

import (
	"net/http"
	"strconv"

	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/recommendation-service/dtos"
	"github.com/vanjmali/spotlite/recommendation-service/services"
)

const LIMIT_DEFAULT = 20
const LIMIT_MAX = 50

type RecommendationHandler struct {
	recommendationService *services.RecommendationService
}

func NewRecommendationHandler(rs *services.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{
		recommendationService: rs,
	}
}

// HandleGetRecommendations handles HTTP GET requests to retrieve personalized recommendations for the authenticated user
func (h *RecommendationHandler) HandleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	// Extract authenticated user ID from context
	userID := middlewares.GetUserIdFromContext(r.Context())
	if userID == "" {
		logging.Warnf(r.Context(), "user ID not found in context")
		_ = respond.Unauthorized(w, respond.ErrorMessage("unauthorized"))
		return
	}

	// Parse limit query parameter
	limit := LIMIT_DEFAULT
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit > 0 && parsedLimit <= LIMIT_MAX {
				limit = parsedLimit
			} else if parsedLimit > LIMIT_MAX {
				_ = respond.BadRequest(w, respond.ErrorMessage("limit cannot exceed "+strconv.Itoa(LIMIT_MAX)))
				return
			}
		}
	}

	// Get recommendations from service
	songs, err := h.recommendationService.GetPersonalizedRecommendations(r.Context(), userID, limit)
	if err != nil {
		logging.Errorf(r.Context(), "failed to get recommendations: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	// Build response
	response := dtos.RecommendationResponseDto{
		Songs:   songs,
		Message: "Recommendations retrieved successfully",
	}

	_ = respond.OkJson(w, response)
}

// HandleGetTrendingSongs handles HTTP GET requests to retrieve trending/popular songs
func (h *RecommendationHandler) HandleGetTrendingSongs(w http.ResponseWriter, r *http.Request) {
	// Parse limit query parameter
	limit := LIMIT_DEFAULT
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit > 0 && parsedLimit <= LIMIT_MAX {
				limit = parsedLimit
			} else if parsedLimit > LIMIT_MAX {
				_ = respond.BadRequest(w, respond.ErrorMessage("limit cannot exceed "+strconv.Itoa(LIMIT_MAX)))
				return
			}
		}
	}

	// Get trending songs from service
	songs, err := h.recommendationService.GetTrendingSongs(r.Context(), limit)
	if err != nil {
		logging.Errorf(r.Context(), "failed to get trending songs: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	// Build response
	response := dtos.RecommendationResponseDto{
		Songs:   songs,
		Message: "Trending songs retrieved successfully",
	}

	_ = respond.OkJson(w, response)
}
