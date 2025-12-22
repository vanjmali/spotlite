package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/respond"
	"github.com/vanjmali/spotlite/common-lib/utils"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/services"
)

var (
	VerificationSuccessUrl = utils.MustGetEnv("SRV_USER_VERIFICATION_SUCCESS_URL")
	VerificationFailureUrl = utils.MustGetEnv("SRV_USER_VERIFICATION_FAILURE_URL")
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

// HandleLogin authenticates user credentials and triggers OTP delivery.
func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("failed to login user: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	if err := respond.Ok(w, "OTP sent to email."); err != nil {
		log.Printf("failed to write login response: %v", err)
	}
}

// HandleRegistration func, handles user registration requests and returns adequate responses.
func (h *UserHandler) HandleRegistration(w http.ResponseWriter, r *http.Request) {
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
		default:
			msg = "An unexpected error has occurred."
		}

		if msg != "" {
			_ = respond.Conflict(w, msg)
			return
		}

		log.Printf("failed to register user: %v", err)
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
		log.Printf("failed to verify login otp: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	token, err := h.s.CreateNewToken(r.Context(), user)
	if err != nil {
		log.Printf("failed to create access token: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	refresh, err := h.rts.IssueRefreshToken(r.Context(), user.ID)
	if err != nil {
		log.Printf("failed to issue refresh token: %v", err)
		_ = respond.InternalServerError(w)
		return
	}

	b := map[string]any{
		"access_token":  token,
		"refresh_token": refresh,
	}

	if err := respond.OkJson(w, b); err != nil {
		log.Printf("failed to write verify login otp response: %v", err)
	}
}

// HandleResendOtp resends the OTP code to the user's email if they have a valid login request
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
		_ = respond.BadRequest(w, "Email not found.")
		return
	case errors.Is(err, services.ErrUserInnactive):
		_ = respond.Unauthorized(w, "User is inactive.")
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

// HandleCheckEmail checks if an email is already registered
func (h *UserHandler) HandleCheckEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	// Create DTO and validate the email parameter
	dto := dtos.CheckEmailDto{Email: email}
	if err := h.v.Struct(dto); err != nil {
		_ = respond.BadRequest(w, "Invalid email format.")
		return
	}

	exists, err := h.s.EmailExists(r.Context(), dto.Email)
	if err != nil {
		_ = respond.InternalServerError(w)
		return
	}

	_ = respond.OkJson(w, map[string]bool{"exists": exists})
}
