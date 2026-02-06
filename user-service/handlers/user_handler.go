package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/telemetry"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/services"
)

// UserHandler wires HTTP handlers to the user service and validators.
type UserHandler struct {
	s      *services.UserService
	v      *validator.Validate
	rts    *services.RefreshTokenService
	config UserHandlerConfig
}

type UserHandlerConfig struct {
	VerificationSuccessUrl string
	VerificationFailureUrl string
}

func NewUserHandler(s services.UserService, v validator.Validate, rts services.RefreshTokenService, c UserHandlerConfig) *UserHandler {
	h := UserHandler{s: &s, v: &v, rts: &rts, config: c}
	return &h
}

func (h *UserHandler) HandleChangePassword(w http.ResponseWriter, r *http.Request) {
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
		_ = respond.BadRequest(w, respond.ErrorMessage("Invalid current password."))
		return

	case errors.Is(err, services.ErrTooFrequentPasswordChange):
		_ = respond.BadRequest(w, respond.ErrorMessage("Password can only be changed once every 24 hours."))
		return
	case err != nil:
		_ = respond.InternalServerError(w)
		return
	}
	if err := respond.Ok(w, "Password changed successfully."); err != nil {
		log.Printf("failed to write change password response: %v", err)
	}
}

// HandleLogin authenticates user credentials and triggers OTP delivery.
func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req dtos.UserLoginDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s failed to process login request: %v", telemetry.TraceID(r.Context()), err)
		}
		return
	}

	err := h.s.Login(r.Context(), &req)

	switch {
	case errors.Is(err, services.ErrUserNotFound):
		_ = respond.Unauthorized(w, respond.ErrorMessage("Invalid credentials."))
		return
	case errors.Is(err, services.ErrBadCredentials):
		_ = respond.Unauthorized(w, respond.ErrorMessage("Invalid credentials."))
		return
	case errors.Is(err, services.ErrExpiredPassword):
		_ = respond.Unauthorized(w, respond.ErrorMessage("Password expired."))
		return
	case errors.Is(err, services.ErrUserInactive):
		_ = respond.Unauthorized(w, respond.ErrorMessage("User is inactive."))
		return
	case err != nil:
		log.Printf("trace_id=%s failed to login user: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.Ok(w, "OTP sent to email."); err != nil {
		log.Printf("trace_id=%s failed to write login response: %v", telemetry.TraceID(r.Context()), err)
	}
}

// HandleRegistration func, handles user registration requests and returns adequate responses.
func (h *UserHandler) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	// trying to decode the registration request dto
	var req dtos.UserRegistrationDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s failed to process registration request: %v", telemetry.TraceID(r.Context()), err)
		}
		return
	}

	// initializes registration after decoding and validation went well
	err := h.s.Register(r.Context(), &req)
	if err != nil {
		var conflict respond.ErrorMessagePayload
		hasConflict := false

		switch {
		case errors.Is(err, services.ErrUsernameTaken):
			conflict = respond.ErrorMessageWithCode("Username is already taken.", "username_taken")
			hasConflict = true
		case errors.Is(err, services.ErrEmailTaken):
			conflict = respond.ErrorMessageWithCode("Email is already taken.", "email_taken")
			hasConflict = true
		}

		if hasConflict {
			_ = respond.Conflict(w, conflict)
			return
		}

		log.Printf("trace_id=%s failed to register user: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}

// HandleAccountVerification func, handles user account verification requests and redirects to success/failure pages
// depending on the result.
func (h *UserHandler) HandleAccountVerification(w http.ResponseWriter, r *http.Request) {
	// fetches token query parameter value
	token := r.URL.Query().Get("token")

	// handle if there is no token sent as query param
	if token == "" {
		http.Redirect(w, r, h.config.VerificationFailureUrl, http.StatusSeeOther)
		return
	}

	// initialize account verification
	err := h.s.VerifyAccount(r.Context(), token)
	if err != nil {
		if errors.Is(err, repositories.ErrTokenExpired) {
			http.Redirect(w, r, h.config.VerificationFailureUrl, http.StatusSeeOther)
			return
		}

		log.Printf("trace_id=%s failed to verify account: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	// handle account verification success
	http.Redirect(w, r, h.config.VerificationSuccessUrl, http.StatusSeeOther)
}

// HandleVerifyLoginOtp validates the OTP and issues a JWT token on success.
func (h *UserHandler) HandleVerifyLoginOtp(w http.ResponseWriter, r *http.Request) {
	var req dtos.VerifyLoginOtpDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("trace_id=%s failed to process verify login otp request: %v", telemetry.TraceID(r.Context()), err)
		}
		return
	}

	user, err := h.s.VerifyLoginOtp(r.Context(), &req)

	if errors.Is(err, services.ErrOtpExpired) || errors.Is(err, services.ErrOtpInvalid) {
		_ = respond.Unauthorized(w, respond.ErrorMessage("Invalid or expired OTP."))
		return
	}

	if err != nil {
		log.Printf("trace_id=%s failed to verify login otp: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	token, err := h.s.CreateNewToken(r.Context(), user)
	if err != nil {
		log.Printf("trace_id=%s failed to create access token: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	refresh, expiresAt, err := h.rts.IssueRefreshToken(r.Context(), user.ID)
	if err != nil {
		log.Printf("trace_id=%s failed to issue refresh token: %v", telemetry.TraceID(r.Context()), err)
		_ = respond.InternalServerError(w)
		return
	}

	setRefreshCookie(w, refresh, expiresAt)

	b := map[string]any{
		"access_token": token,
	}

	if err := respond.OkJson(w, b); err != nil {
		log.Printf("trace_id=%s failed to write verify login otp response: %v", telemetry.TraceID(r.Context()), err)
	}
}

// HandleResendOtp resends the OTP code to the user's email if they have a valid login request.
func (h *UserHandler) HandleResendOtp(w http.ResponseWriter, r *http.Request) {
	var req dtos.ResendOtpDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			log.Printf("failed to process resend otp request: %v", err)
		}
		return
	}

	err := h.s.ResendLoginOtp(r.Context(), req.Email)

	switch {
	case errors.Is(err, services.ErrUserNotFound):
		_ = respond.BadRequest(w, respond.ErrorMessage("Email not found."))
		return
	case errors.Is(err, services.ErrUserInactive):
		_ = respond.Unauthorized(w, respond.ErrorMessage("User is inactive."))
		return
	case err != nil:
		log.Printf("failed to resend otp: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.Ok(w, "OTP resent to email."); err != nil {
		log.Printf("failed to write resend otp response: %v", err)
	}
}

// HandleCheckEmail checks if an email is already registered.
func (h *UserHandler) HandleCheckEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		_ = respond.BadRequest(w, respond.ErrorMessage("Email is required"))
		return
	}

	// validate email format
	if err := h.v.Var(email, "required,email"); err != nil {
		_ = respond.BadRequest(w, respond.ErrorMessage("Invalid email"))
		return
	}

	exists, err := h.s.EmailExists(r.Context(), email)
	if err != nil {
		_ = respond.InternalServerError(w)
		return
	}

	_ = respond.OkJson(w, map[string]bool{"exists": exists})
}

// HandleLogout revokes the refresh token (if present) and clears the cookie.
func (h *UserHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName())
	if err == nil && cookie.Value != "" {
		if err := h.rts.RevokeRefreshToken(r.Context(), cookie.Value); err != nil &&
			!errors.Is(err, services.ErrRefreshInvalid) {
			log.Printf("trace_id=%s failed to revoke refresh token: %v", telemetry.TraceID(r.Context()), err)
		}
	}

	clearRefreshCookie(w)
	respond.NoContent(w)
}
