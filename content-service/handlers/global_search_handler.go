package handlers

import (
	"log"
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/content/services"
)

// GlobalSearchHandler handles global search requests across multiple content types.
type GlobalSearchHandler struct {
	s *services.GlobalSearchService
}

func NewGlobalSearchHandler(s *services.GlobalSearchService) *GlobalSearchHandler {
	return &GlobalSearchHandler{
		s: s,
	}
}

func (h *GlobalSearchHandler) HandleGlobalSearch(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		log.Printf("trace_id=%s search query 'q' is required", telemetry.TraceID(r.Context()))
		_ = respond.BadRequest(w, respond.ErrorMessage("Search query 'q' is required"))
		return
	}

	result, err := h.s.GetGlobalSearch(r.Context(), searchTerm)
	if err != nil {
		log.Printf("trace_id=%s failed to perform global search: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, result); err != nil {
		log.Printf("trace_id=%s failed to write global search response: %v", telemetry.TraceID(r.Context()), err)
	}
}
