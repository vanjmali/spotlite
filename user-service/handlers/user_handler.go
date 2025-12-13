package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/entities"
	"github.com/vanjmali/spotlite/user-service/repositories"
	"github.com/vanjmali/spotlite/user-service/services"
	"github.com/vanjmali/spotlite/user-service/validation"
)

const (
	VerificationSuccessUrl = "https://google.com"
	VerificationFailureUrl = "https://apple.com"
)

type UserHandler struct {
	s *services.UserService
	v *validator.Validate
}

func NewUserHandler(s services.UserService, v validator.Validate) *UserHandler {
	h := UserHandler{s: &s, v: &v}
	return &h
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	errorResponse := entities.ErrorResponse{Status: statusCode, Message: message}
	json.NewEncoder(w).Encode(errorResponse)
}

// validateUserRegistration func, validates registration request dto field values,
func (h *UserHandler) validateUserRegistration(dto *dtos.UserRegistrationDto) error {
	return h.v.Struct(dto)
}

// HandleRegistration func, handles user registration requests and returns adequate responses
func (h *UserHandler) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// trying to decode the registration request dto
	var req dtos.UserRegistrationDto
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// validates request field values
	if err := h.validateUserRegistration(&req); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal validation error"})
			return
		}

		var errors []entities.FieldError

		for _, err := range err.(validator.ValidationErrors) {

			errors = append(errors, entities.FieldError{
				Field:   strings.ToLower(err.Field()),
				Message: validation.GetErrorMsg(err),
			})
		}

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"errors": errors,
		})
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

/*
HandleAccountVerification func, handles user account verification requests and redirects to success/failure pages
depending on the result
*/
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
		case errors.Is(err, repositories.TokenExpiredErr):
			http.Redirect(w, r, VerificationFailureUrl, http.StatusSeeOther)
			return
		default:
			http.Redirect(w, r, VerificationFailureUrl, http.StatusSeeOther)
			return
		}
	}

	// handle account verification success
	http.Redirect(w, r, VerificationSuccessUrl, http.StatusSeeOther)
	return
}
