package respond

import (
	"net/http"
)

// ErrorResponse represents a generic JSON error payload returned by the API.
type ErrorResponse struct {
	HttpCode int `json:"-"`
	// Code is a machine-readable error code. Use a snake_case string.
	Code string `json:"code"`
	// Message provides a human-readable description of the error.
	// If empty, it defaults to the standard HTTP status text.
	Message string `json:"message"`
	// Fields provides additional field-level error details, if applicable.
	// It is omitted if there are no field errors.
	Fields map[string]string `json:"fields,omitempty"`
}

// Error issues an error response with the specified HTTP status code and message.
func Error(w http.ResponseWriter, e ErrorResponse) error {
	r := ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		Fields:  e.Fields,
	}

	if e.Message == "" {
		e.Message = http.StatusText(e.HttpCode)
	}

	return writeJson(w, e.HttpCode, r)
}

func TooManyRequests(w http.ResponseWriter) error {
	r := ErrorResponse{
		HttpCode: http.StatusTooManyRequests,
		Code:     "rate_limit_exceeded",
		Message:  "Too many requests. Please wait and retry.",
	}

	return Error(w, r)
}

// ValidationError issues a validation error response with field-level details.
// It uses HTTP status code 400 (Bad Request) with code "validation_error" and a generic message.
func ValidationError(w http.ResponseWriter, fields map[string]string) error {
	r := ErrorResponse{
		HttpCode: http.StatusBadRequest,
		Code:     "validation_error",
		Message:  "One or more fields have validation errors.",
		Fields:   fields,
	}

	return Error(w, r)
}

// NotFound issues a 404 Not Found error response with a standard message.
func NotFound(w http.ResponseWriter) error {
	r := ErrorResponse{
		HttpCode: http.StatusNotFound,
		Code:     "not_found",
		Message:  "The requested resource was not found.",
	}

	return Error(w, r)
}

// InternalServerError issues a 500 Internal Server Error response with a standard message.
func InternalServerError(w http.ResponseWriter) error {
	r := ErrorResponse{
		HttpCode: http.StatusInternalServerError,
		Code:     "internal_server_error",
		Message:  "An unexpected error has occurred.",
	}

	return Error(w, r)
}

// BadRequest issues a 400 Bad Request error response with a custom message.
// Messages should be clear and concise to help clients understand the issue. Use capitalization and punctuation appropriately.
func BadRequest(w http.ResponseWriter, message ...string) error {
	msg := "Bad request."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	r := ErrorResponse{
		HttpCode: http.StatusBadRequest,
		Code:     "bad_request",
		Message:  msg,
	}

	return Error(w, r)
}

func UnprocessableEntity(w http.ResponseWriter, message ...string) error {
	msg := "Unprocessable entity"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	r := ErrorResponse{
		HttpCode: http.StatusUnprocessableEntity,
		Code:     "unprocessable_entity",
		Message:  msg,
	}

	return Error(w, r)
}

// Unauthorized issues a 401 Unauthorized error response.
// If a custom message is provided, it uses that; otherwise it falls back to the standard message.
func Unauthorized(w http.ResponseWriter, message ...string) error {
	msg := "You are not authorized to access this resource."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	r := ErrorResponse{
		HttpCode: http.StatusUnauthorized,
		Code:     "unauthorized",
		Message:  msg,
	}

	return Error(w, r)
}

// Forbidden issues a 403 Forbidden error response with a standard message.
func Forbidden(w http.ResponseWriter) error {
	r := ErrorResponse{
		HttpCode: http.StatusForbidden,
		Code:     "forbidden",
		Message:  "You do not have permission to access this resource.",
	}

	return Error(w, r)
}

func Conflict(w http.ResponseWriter, message ...string) error {
	msg := "Conflict occurred."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	r := ErrorResponse{
		HttpCode: http.StatusConflict,
		Code:     "conflict",
		Message:  msg,
	}

	return Error(w, r)
}

// NotImplemented issues a 501 Not Implemented error response with a standard message.
func NotImplemented(w http.ResponseWriter, message ...string) error {
	msg := "Not implemented."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	r := ErrorResponse{
		HttpCode: http.StatusNotImplemented,
		Code:     "not_implemented",
		Message:  msg,
	}

	return Error(w, r)
}
