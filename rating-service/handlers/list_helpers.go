package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/rating-service/services"
)

// handleListResponse is a helper function to handle list responses for different entities.
func handleListResponse(w http.ResponseWriter, r *http.Request, logLabel string, fetch func(context.Context) (any, error)) {
	resp, err := fetch(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, respond.ErrorMessage("Invalid ID format"))
		default:
			logging.Errorf(r.Context(), "failed to list %s: %v", logLabel, err)
			_ = respond.InternalServerError(w)
		}
		return
	}

	if err := respond.OkJson(w, resp); err != nil {
		logging.Errorf(r.Context(), "failed to write list %s response: %v", logLabel, err)
	}
}
