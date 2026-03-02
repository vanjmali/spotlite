package handlers

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/analytics-service/mappers"
	"github.com/vanjmali/spotlite/analytics-service/services"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/middlewares"
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

// HandleGetUserAnalytics handles GET /
// Returns analytics summary for the currently authenticated user.
func (h *AnalyticsHandler) HandleGetUserAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middlewares.GetUserIdFromContext(ctx)

	if userID == "" {
		_ = respond.Unauthorized(w)
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
