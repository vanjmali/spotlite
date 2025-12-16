package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/services"
)

type RefreshTokenHandler struct {
	s  *services.RefreshTokenService
	us *services.UserService
}

func NewRefreshTokenHandler(rts services.RefreshTokenService, us services.UserService) *RefreshTokenHandler {
	return &RefreshTokenHandler{s: &rts, us: &us}
}

func (h *RefreshTokenHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dtos.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		sendErrorResponse(w, http.StatusBadRequest, "invalid payload")
		return
	}

	userID, err := h.s.GetRefreshTokenId(r.Context(), req.RefreshToken)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.us.FindByID(r.Context(), userID)
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	access, err := h.us.CreateNewToken(r.Context(), user)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "an unexpected error has occurred")
		return
	}

	json.NewEncoder(w).Encode(dtos.RefreshResponse{
		AccessToken: access,
	})
}
