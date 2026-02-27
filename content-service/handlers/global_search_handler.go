package handlers

import (
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/respond"
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
		logging.Warnf(r.Context(), "search query 'q' is required")
		_ = respond.BadRequest(w, respond.ErrorMessage("Search query 'q' is required"))
		return
	}

	result, err := h.s.GetGlobalSearch(r.Context(), searchTerm)
	if err != nil {
		logging.Errorf(r.Context(), "failed to perform global search: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	getSongRatings(r.Context(), result.Songs)
	for i := range result.Albums {
		getSongRatings(r.Context(), result.Albums[i].Songs)
	}

	if err := respond.OkJson(w, result); err != nil {
		logging.Errorf(r.Context(), "failed to write global search response: %v", err)
	}
}
