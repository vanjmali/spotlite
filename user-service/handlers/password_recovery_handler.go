package handlers

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/logging"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/services"
)

// PasswordRecoveryHandler handles all password recovery operations.
type PasswordRecoveryHandler struct {
	prs *services.PasswordRecoveryService
	v   *validator.Validate
}

func NewPasswordRecoveryHandler(prs services.PasswordRecoveryService, v validator.Validate) *PasswordRecoveryHandler {
	return &PasswordRecoveryHandler{prs: &prs, v: &v}
}

// HandleRequestPasswordReset processes password reset requests and sends a magic link email.
func (h *PasswordRecoveryHandler) HandleRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req dtos.RequestPasswordResetDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "failed to process password reset request: %v", err)
		}
		return
	}

	err := h.prs.RequestPasswordReset(r.Context(), req.Email)
	if err != nil {
		logging.Errorf(r.Context(), "failed to request password reset: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	// Always return success to avoid email enumeration
	_ = respond.Ok(w, "If email exists, recovery link will be sent")
}

// HandleValidateRecoveryToken validates that a password recovery token is valid and not expired.
func (h *PasswordRecoveryHandler) HandleValidateRecoveryToken(w http.ResponseWriter, r *http.Request) {
	var req dtos.ValidateRecoveryTokenDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "failed to process token validation request: %v", err)
		}
		return
	}

	_, err := h.prs.ValidateRecoveryToken(r.Context(), req.Token)

	switch {
	case errors.Is(err, services.ErrInvalidRecoveryToken):
		_ = respond.Unauthorized(w, respond.ErrorMessage("Invalid recovery token"))
		return
	case errors.Is(err, services.ErrRecoveryTokenExpired):
		_ = respond.Unauthorized(w, respond.ErrorMessage("Recovery token expired"))
		return
	case errors.Is(err, services.ErrRecoveryTokenUsed):
		_ = respond.Unauthorized(w, respond.ErrorMessage("Recovery token already used"))
		return
	case err != nil:
		logging.Errorf(r.Context(), "failed to validate recovery token: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	_ = respond.Ok(w, "Token is valid")
}

// HandleResetPassword resets the user's password using a valid recovery token.
func (h *PasswordRecoveryHandler) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dtos.ResetPasswordDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "failed to process reset password request: %v", err)
		}
		return
	}

	err := h.prs.ResetPassword(r.Context(), &req)

	switch {
	case errors.Is(err, services.ErrInvalidRecoveryToken):
		_ = respond.Unauthorized(w, respond.ErrorMessage("Invalid recovery token"))
		return
	case errors.Is(err, services.ErrRecoveryTokenExpired):
		_ = respond.Unauthorized(w, respond.ErrorMessage("Recovery token expired"))
		return
	case errors.Is(err, services.ErrRecoveryTokenUsed):
		_ = respond.Unauthorized(w, respond.ErrorMessage("Recovery token already used"))
		return
	case errors.Is(err, services.ErrUserNotFound):
		_ = respond.NotFound(w)
		return
	case err != nil:
		logging.Errorf(r.Context(), "failed to reset password: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	_ = respond.Ok(w, "Password reset successfully")
}
