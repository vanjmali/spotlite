package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/services"
)

// handleListResponse is a helper function to handle list responses for different entities.
func handleListResponse(w http.ResponseWriter, r *http.Request, logLabel string, fetch func(context.Context) (any, error)) {
	resp, err := fetch(r.Context())
	if err != nil {
		switch {
		case errors.Is(err, services.ErrObjectIdCastFailed):
			_ = respond.BadRequest(w, respond.ErrorMessage("Invalid ID format"))
		default:
			log.Printf("trace_id=%s failed to list %s: %v", telemetry.TraceID(r.Context()), logLabel, err)
			_ = respond.InternalServerError(w)
		}
		return
	}

	if err := respond.OkJson(w, resp); err != nil {
		log.Printf("trace_id=%s failed to write list %s response: %v", telemetry.TraceID(r.Context()), logLabel, err)
	}
}
