package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/user-service/dtos"
	"github.com/vanjmali/spotlite/user-service/models"
	"github.com/vanjmali/spotlite/user-service/services"
	"github.com/vanjmali/spotlite/user-service/validation"
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
	errorResponse := models.ErrorResponse{Status: statusCode, Message: message}
	json.NewEncoder(w).Encode(errorResponse)
}

func (h *UserHandler) validateUserRegistration(dto *dtos.UserRegistrationDto) error {
	return h.v.Struct(dto)
}

func (h *UserHandler) HandleRegistration(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req dtos.UserRegistrationDto
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.validateUserRegistration(&req); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Internal validation error"})
			return
		}

		var errors []models.FieldError

		for _, err := range err.(validator.ValidationErrors) {

			errors = append(errors, models.FieldError{
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

	err = h.s.Register(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUsernameTaken):
			sendErrorResponse(w, http.StatusConflict, err.Error())
			return
		case errors.Is(err, services.ErrEmailTaken):
			sendErrorResponse(w, http.StatusConflict, err.Error())
			return
		default:
			sendErrorResponse(w, http.StatusInternalServerError, "an unexpected error has occurred")
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
	return
}
