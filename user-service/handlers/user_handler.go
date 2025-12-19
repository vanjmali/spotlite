package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/services"
)

var (
	VerificationSuccessUrl = utils.MustGetEnv("APP_VERIFICATION_SUCCESS_URL")
	VerificationFailureUrl = utils.MustGetEnv("APP_VERIFICATION_FAILURE_URL")
)

// UserHandler wires HTTP handlers to the user service and validators.
type UserHandler struct {
	s   *services.UserService
	v   *validator.Validate
	rts *services.RefreshTokenService
}

func NewUserHandler(s services.UserService, v validator.Validate, rts services.RefreshTokenService) *UserHandler {
	h := UserHandler{s: &s, v: &v, rts: &rts}
	return &h
}

func (h *UserHandler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dtos.ChangePasswordDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("failed to process change password request: %v", err)
		}
		return
	}

	err := h.s.ChangePassword(r.Context(), &req)

	switch {
	case errors.Is(err, services.ErrInvalidCurrentPassword):
		_ = respond.Unauthorized(w, "Invalid current password.")
		return

	case errors.Is(err, services.ErrTooFrequentPasswordChange):
		_ = respond.BadRequest(w, "Password changed too frequently.")
		return
	case err != nil:
		_ = respond.InternalServerError(w)
		return
	}
	w.WriteHeader(http.StatusOK)

	b := map[string]any{
		"message": "Password changed successfully",
	}

	if err := json.NewEncoder(w).Encode(b); err != nil {
		_ = respond.InternalServerError(w)
		return
	}
}

// HandleLogin authenticates user credentials and triggers OTP delivery.
func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dtos.UserLoginDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("failed to process login request: %v", err)
		}
		return
	}

	err := h.s.Login(r.Context(), &req)

	switch {
	case errors.Is(err, services.ErrBadCredentials):
		_ = respond.Unauthorized(w, "Invalid credentials.")
		return
	case errors.Is(err, services.ErrExpiredPassword):
		_ = respond.Unauthorized(w, "Password expired.")
		return
	case errors.Is(err, services.ErrUserInnactive):
		_ = respond.Unauthorized(w, "User is inactive.")
		return
	case err != nil:
		_ = respond.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]any{
		"otp_required": true,
		"message":      "OTP sent to email",
	}); err != nil {
		log.Printf("failed to write login response: %v", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

// HandleRegistration func, handles user registration requests and returns adequate responses.
func (h *UserHandler) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// trying to decode the registration request dto
	var req dtos.UserRegistrationDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("failed to process registration request: %v", err)
		}
		return
	}

	// initializes registration after decoding and validation went well
	err := h.s.Register(r.Context(), &req)
	if err != nil {
		var msg string
		switch {
		case errors.Is(err, services.ErrUsernameTaken):
			msg = "Username is already taken."
		case errors.Is(err, services.ErrEmailTaken):
			msg = "Email is already taken."
		}

		if msg != "" {
			_ = respond.Conflict(w, msg)
			return
		}

		_ = respond.InternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleAccountVerification func, handles user account verification requests and redirects to success/failure pages
// depending on the result.
func (h *UserHandler) HandleAccountVerification(w http.ResponseWriter, r *http.Request) {
	// fetches token query parameter value
	token := r.URL.Query().Get("token")

	// handle if there is no token sent as query param
	if token == "" {
		http.Redirect(w, r, VerificationFailureUrl, http.StatusSeeOther)
		return
	}

	// initialize account verification
	err := h.s.VerifyAccount(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrTokenExpired):
			http.Redirect(w, r, VerificationFailureUrl, http.StatusSeeOther)
			return
		default:
			http.Redirect(w, r, VerificationFailureUrl, http.StatusSeeOther)
			return
		}
	}

	// handle account verification success
	http.Redirect(w, r, VerificationSuccessUrl, http.StatusSeeOther)
}

// HandleVerifyLoginOtp validates the OTP and issues a JWT token on success.
func (h *UserHandler) HandleVerifyLoginOtp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dtos.VerifyLoginOtpDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("failed to process verify login otp request: %v", err)
		}
		return
	}

	user, err := h.s.VerifyLoginOtp(r.Context(), &req)

	if errors.Is(err, services.ErrOtpExpired) || errors.Is(err, services.ErrOtpInvalid) {
		_ = respond.Unauthorized(w, "Invalid or expired OTP.")
		return
	}

	if err != nil {
		_ = respond.InternalServerError(w)
		return
	}

	token, err := h.s.CreateNewToken(r.Context(), user)
	if err != nil {
		_ = respond.InternalServerError(w)
		return
	}

	refresh, err := h.rts.IssueRefreshToken(r.Context(), user.ID)
	if err != nil {
		_ = respond.InternalServerError(w)
		return
	}

	b := map[string]any{
		"access_token":  token,
		"refresh_token": refresh,
	}

	if err := json.NewEncoder(w).Encode(b); err != nil {
		_ = respond.InternalServerError(w)
	}
}
