package handlers

import (
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/services"
)

type RefreshTokenHandler struct {
	s  *services.RefreshTokenService
	us *services.UserService
	v  *validator.Validate
}

func NewRefreshTokenHandler(rts services.RefreshTokenService, us services.UserService, v validator.Validate) *RefreshTokenHandler {
	return &RefreshTokenHandler{s: &rts, us: &us, v: &v}
}

func (h *RefreshTokenHandler) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dtos.RefreshRequest
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s failed to process refresh token request: %v", telemetry.TraceID(r.Context()), err)
		}
		return
	}

	userID, err := h.s.GetRefreshTokenId(r.Context(), req.RefreshToken)
	if err != nil {
		_ = respond.Unauthorized(w)
		return
	}

	user, err := h.us.FindByID(r.Context(), userID)
	if err != nil {
		_ = respond.Unauthorized(w)
		return
	}

	access, err := h.us.CreateNewToken(r.Context(), user)
	if err != nil {
		log.Printf("trace_id=%s failed to create new access token: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	b := dtos.RefreshResponse{
		AccessToken: access,
	}

	if err := respond.OkJson(w, b); err != nil {
		log.Printf("trace_id=%s failed to write refresh token response: %v", telemetry.TraceID(r.Context()), err)
	}
}
