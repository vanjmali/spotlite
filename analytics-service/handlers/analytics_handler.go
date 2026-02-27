package handlers

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/analytics-service/mappers"
	"github.com/vanjmali/spotlite/analytics-service/services"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/respond"
)

// AnalyticsHandler handles HTTP requests for analytics queries
type AnalyticsHandler struct {
	service   *services.AnalyticsService
	validator *validator.Validate
}

// NewAnalyticsHandler creates a new AnalyticsHandler instance
func NewAnalyticsHandler(s *services.AnalyticsService, v *validator.Validate) *AnalyticsHandler {
	return &AnalyticsHandler{
		service:   s,
		validator: v,
	}
}

// HandleGetUserAnalytics handles GET /analytics/{userID}
// Returns analytics summary for a user (total plays, ratings, top artists, etc.)
func (h *AnalyticsHandler) HandleGetUserAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		_ = respond.BadRequest(w, respond.ErrorMessage("userID is required"))
		return
	}

	analytics, err := h.service.GetUserAnalytics(ctx, userID)
	if err != nil {
		if errors.Is(err, services.ErrAnalyticsNotFound) {
			logging.Warnf(ctx, "analytics not found for user: %s", userID)
			_ = respond.NotFound(w)
			return
		}
		logging.Errorf(ctx, "failed to get user analytics: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	responseDto := mappers.ToUserAnalyticsResponseDto(analytics)
	_ = respond.OkJson(w, responseDto)
}

// HandleGetActivityHistory handles GET /activity-history/{userID}
// Returns activity history timeline for a user
func (h *AnalyticsHandler) HandleGetActivityHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		_ = respond.BadRequest(w, respond.ErrorMessage("userID is required"))
		return
	}

	history, err := h.service.GetUserActivityHistory(ctx, userID)
	if err != nil {
		if errors.Is(err, services.ErrActivityHistoryNotFound) {
			logging.Warnf(ctx, "activity history not found for user: %s", userID)
			_ = respond.NotFound(w)
			return
		}
		logging.Errorf(ctx, "failed to get activity history: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	responseDto := mappers.ToUserActivityHistoryResponseDto(history)
	_ = respond.OkJson(w, responseDto)
}
