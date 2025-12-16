package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/services"
	"github.com/vanjmali/spotlite/user-service/utils"
	"github.com/vanjmali/spotlite/user-service/validation"
)

var (
	// VerificationSuccessUrl redirects the user after a successful account verification.
	VerificationSuccessUrl = utils.MustGetEnv("APP_VERIFICATION_SUCCESS_URL")
	// VerificationFailureUrl redirects the user when verification fails.
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

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	errorResponse := entities.ErrorResponse{Status: statusCode, Message: message}
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		log.Printf("failed to write error response: %v", err)
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

// validateUserRegistration func, validates registration request dto field values,.
func (h *UserHandler) validateUserRegistration(dto *dtos.UserRegistrationDto) error {
	return h.v.Struct(dto)
}

// HandleLogin authenticates user credentials and triggers OTP delivery.
func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dtos.UserLoginDto
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid payload")
		return
	}

	err = h.s.Login(r.Context(), &req)

	switch {
	case errors.Is(err, services.ErrBadCredentials):
		sendErrorResponse(w, http.StatusUnauthorized, "invalid credentials")
		return
	case errors.Is(err, services.ErrExpiredPassword):
		sendErrorResponse(w, http.StatusUnauthorized, "password expired")
		return
	case errors.Is(err, services.ErrUserInnactive):
		sendErrorResponse(w, http.StatusUnauthorized, "user is inactive")
		return
	case err != nil:
		sendErrorResponse(w, http.StatusInternalServerError, "an unexpected error has occurred")
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
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		if err == io.EOF {
			sendErrorResponse(w, http.StatusBadRequest, "request body can't be empty")
			return
		}

		syntaxError := &json.SyntaxError{}
		if errors.As(err, &syntaxError) {
			sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON format: %s", err))
			return
		}

		sendErrorResponse(
			w,
			http.StatusInternalServerError,
			"an unexpected error has occurred while processing your request",
		)
		return
	}

	// validates request field values
	if err := h.validateUserRegistration(&req); err != nil {
		invalidValidationError := &validator.InvalidValidationError{}
		if errors.As(err, &invalidValidationError) {
			sendErrorResponse(w, http.StatusInternalServerError, "internal validation error")
			return
		}

		var fieldErrors []entities.FieldError

		var validationErrs validator.ValidationErrors
		_ = errors.As(err, &validationErrs)

		for _, verr := range validationErrs {
			fieldErrors = append(fieldErrors, entities.FieldError{
				Field:   strings.ToLower(verr.Field()),
				Message: validation.GetErrorMsg(verr),
			})
		}

		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": fieldErrors,
		}); err != nil {
			log.Printf("failed to write validation errors: %v", err)
			http.Error(w, "failed to write response", http.StatusInternalServerError)
		}
		return
	}

	// initializes registration after decoding and validation went well
	err = h.s.Register(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUsernameTaken) || errors.Is(err, services.ErrEmailTaken):
			sendErrorResponse(w, http.StatusConflict, err.Error())
			return
		default:
			sendErrorResponse(w, http.StatusInternalServerError, "an unexpected error has occurred")
			return
		}
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "invalid payload")
		return
	}

	user, err := h.s.VerifyLoginOtp(r.Context(), &req)

	if errors.Is(err, services.ErrOtpExpired) || errors.Is(err, services.ErrOtpInvalid) {
		sendErrorResponse(w, http.StatusUnauthorized, "unauthorized request")
		return
	}

	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "an unexpected error has occurred")
		return
	}

	token, err := h.s.CreateNewToken(r.Context(), user)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "an unexpected error has occurred")
		return
	}

	refresh, err := h.rts.IssueRefreshToken(r.Context(), user.ID)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "an expected error has occurred")
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"access_token":  token,
		"refresh_token": refresh,
	})
}
