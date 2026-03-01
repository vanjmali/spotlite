package handlers

import (
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/pagination"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/user-service/services"
)

type UserActivityHandler struct {
	s *services.UserActivityService
}

func NewUserActivityHandler(s *services.UserActivityService) *UserActivityHandler {
	return &UserActivityHandler{s: s}
}

func (h *UserActivityHandler) HandleGetMyActivities(w http.ResponseWriter, r *http.Request) {
	p := pagination.ParsePagination(r.URL.Query())

	response, err := h.s.ListMyActivities(r.Context(), p.Page, p.Size)
	if err != nil {
		_ = respond.InternalServerError(w)
		return
	}

	_ = respond.OkJson(w, response)
}
