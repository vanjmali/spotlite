package entities

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/user-service/validation"
)

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	errorResponse := ErrorResponse{Status: statusCode, Message: message}
	json.NewEncoder(w).Encode(errorResponse)
}

// ErrorResponse represents a generic JSON error payload returned by the API.
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// FieldError provides field-level validation feedback for client requests.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Validate(w http.ResponseWriter, v *validator.Validate, s any) bool {
	if err := v.Struct(s); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			sendErrorResponse(w, http.StatusInternalServerError, "internal validation error")
			return false
		}

		var errors []FieldError

		for _, err := range err.(validator.ValidationErrors) {

			errors = append(errors, FieldError{
				Field:   strings.ToLower(err.Field()),
				Message: validation.GetErrorMsg(err),
			})
		}

		w.WriteHeader(http.StatusBadRequest)
		b := map[string]any{
			"errors": errors,
		}
		if err := json.NewEncoder(w).Encode(b); err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "internal server error")
			return false
		}

	}
	return true
}
