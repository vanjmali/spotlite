package handlers

import (
	"errors"
	"net/http"

	"github.com/vanjmali/spotlite/common-lib/logging"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/services"
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
	var req dtos.ChangePasswordDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "failed to process change password request: %v", err)
		}
		return
	}

	err := h.s.ChangePassword(r.Context(), &req)

	switch {
	case errors.Is(err, services.ErrInvalidCurrentPassword):
		payload := respond.ErrorMessageWithCode(
			"Invalid current password.",
			"invalid_current_password",
		)
		_ = respond.BadRequest(w, payload)
		return
	case errors.Is(err, services.ErrNewPasswordMatchesCurrent):
		payload := respond.ErrorMessageWithCode(
			"New password must be different from current password.",
			"password_same_as_current",
		)
		_ = respond.BadRequest(w, payload)
		return

	case errors.Is(err, services.ErrTooFrequentPasswordChange):
		payload := respond.ErrorMessageWithCode(
			"Password can only be changed once every 24 hours.",
			"too_frequent_password_change",
		)
		_ = respond.BadRequest(w, payload)
		return
	case err != nil:
		_ = respond.InternalServerError(w)
		return
	}
	if err := respond.Ok(w, "Password changed successfully."); err != nil {
		logging.Errorf(r.Context(), "failed to write change password response: %v", err)
	}
}

// HandleLogin authenticates user credentials and triggers OTP delivery.
func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req dtos.UserLoginDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Errorf(r.Context(), "failed to process login request: %v", err)
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
	case errors.Is(err, services.ErrVerificationRequired):
		_ = respond.Forbidden(
			w,
			respond.ErrorMessageWithCode(
				"Verification email sent. Please check your inbox.",
				"verification_required",
			),
		)
		return
	case err != nil:
		logging.Errorf(r.Context(), "failed to login user: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.Ok(w, "OTP sent to email."); err != nil {
		logging.Errorf(r.Context(), "failed to write login response: %v", err)
	}
}

// HandleRegistration func, handles user registration requests and returns adequate responses.
func (h *UserHandler) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	// trying to decode the registration request dto
	var req dtos.UserRegistrationDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Errorf(r.Context(), "failed to process registration request: %v", err)
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

		logging.Errorf(r.Context(), "failed to register user: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}

// HandleAccountVerification func, handles user account verification requests and redirects to success/failure pages
// depending on the result.
func (h *UserHandler) HandleAccountVerification(w http.ResponseWriter, r *http.Request) {
	var req dtos.VerifyAccountDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Errorf(r.Context(), "failed to process verification request: %v", err)
		}
		return
	}

	// initialize account verification
	err := h.s.VerifyAccount(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, repositories.ErrTokenExpired) {
			m := respond.ErrorMessageWithCode("Invalid or expired verification token.", "verification_failed")
			_ = respond.BadRequest(w, m)
			return
		}

		logging.Errorf(r.Context(), "failed to verify account: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	// handle account verification success
	if err := respond.Ok(w, "Account verified."); err != nil {
		logging.Errorf(r.Context(), "failed to write verify account response: %v", err)
	}
}

// HandleVerifyLoginOtp validates the OTP and issues a JWT token on success.
func (h *UserHandler) HandleVerifyLoginOtp(w http.ResponseWriter, r *http.Request) {
	var req dtos.VerifyLoginOtpDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Errorf(r.Context(), "failed to process verify login otp request: %v", err)
		}
		return
	}

	user, err := h.s.VerifyLoginOtp(r.Context(), &req)

	if errors.Is(err, services.ErrOtpExpired) || errors.Is(err, services.ErrOtpInvalid) {
		_ = respond.Unauthorized(w, respond.ErrorMessage("Invalid or expired OTP."))
		return
	}

	if err != nil {
		logging.Errorf(r.Context(), "failed to verify login otp: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	token, err := h.s.CreateNewToken(r.Context(), user)
	if err != nil {
		logging.Errorf(r.Context(), "failed to create access token: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	refresh, expiresAt, err := h.rts.IssueRefreshToken(r.Context(), user.ID)
	if err != nil {
		logging.Errorf(r.Context(), "failed to issue refresh token: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	setRefreshCookie(w, refresh, expiresAt)

	b := map[string]any{
		"access_token": token,
	}

	if err := respond.OkJson(w, b); err != nil {
		logging.Errorf(r.Context(), "failed to write verify login otp response: %v", err)
	}
}

// HandleResendOtp resends the OTP code to the user's email if they have a valid login request.
func (h *UserHandler) HandleResendOtp(w http.ResponseWriter, r *http.Request) {
	var req dtos.ResendOtpDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Warnf(r.Context(), "failed to process resend otp request: %v", err)
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
		logging.Errorf(r.Context(), "failed to resend otp: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.Ok(w, "OTP resent to email."); err != nil {
		logging.Errorf(r.Context(), "failed to write resend otp response: %v", err)
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

// HandleCheckUsername checks if a username is already registered.
func (h *UserHandler) HandleCheckUsername(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		_ = respond.BadRequest(w, respond.ErrorMessage("Username is required"))
		return
	}

	if err := h.v.Var(username, "required,validusername"); err != nil {
		_ = respond.BadRequest(w, respond.ErrorMessage("Invalid username"))
		return
	}

	exists, err := h.s.UsernameExists(r.Context(), username)
	if err != nil {
		_ = respond.InternalServerError(w)
		return
	}

	_ = respond.OkJson(w, map[string]bool{"exists": exists})
}

// HandleGetProfile returns profile info for authenticated user.
func (h *UserHandler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.s.GetProfile(r.Context())
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid user id"))
		return
	case errors.Is(err, services.ErrUserNotFound):
		_ = respond.NotFound(w)
		return
	case err != nil:
		logging.Errorf(r.Context(), "failed to get profile: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.OkJson(w, profile); err != nil {
		logging.Errorf(r.Context(), "failed to write profile response: %v", err)
	}
}

// HandleUpdateProfile updates profile info for authenticated user.
func (h *UserHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req dtos.UpdateProfileDto
	if ok, err := requests.ReadAndValidateJson(w, h.v, r.Body, &req); !ok {
		if err != nil {
			logging.Errorf(r.Context(), "failed to process update profile request: %v", err)
		}
		return
	}

	err := h.s.UpdateProfile(r.Context(), &req)
	switch {
	case errors.Is(err, services.ErrObjectIdCastFailed):
		_ = respond.BadRequest(w, respond.ErrorMessage("invalid user id"))
		return
	case errors.Is(err, services.ErrUserNotFound):
		_ = respond.NotFound(w)
		return
	case errors.Is(err, services.ErrUsernameTaken):
		_ = respond.Conflict(w, respond.ErrorMessageWithCode("Username is already taken.", "username_taken"))
		return
	case err != nil:
		logging.Errorf(r.Context(), "failed to update profile: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	respond.NoContent(w)
}

// HandleLogout revokes the refresh token (if present) and clears the cookie.
func (h *UserHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName())
	if err == nil && cookie.Value != "" {
		if err := h.rts.RevokeRefreshToken(r.Context(), cookie.Value); err != nil &&
			!errors.Is(err, services.ErrRefreshInvalid) {
			logging.Errorf(r.Context(), "failed to revoke refresh token: %v", err)
		}
	}

	clearRefreshCookie(w)
	respond.NoContent(w)
}
