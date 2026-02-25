package respond

import (
	"net/http"
)

// ErrorMessagePayload represents a simple error message structure.
// If any field is omitted, it defaults to generic message depending on the HTTP status code.
//
// Use respond.ErrorMessage() or respond.ErrorMessageWithCode() to create instances of this struct with default values.
type ErrorMessagePayload struct {
	// Message is a human-readable description of the error.
	Message string
	// Code is a machine-readable error code. Use a snake_case string.
	Code string
}

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

// ErrorMessage is a helper function to create an ErrorMessagePayload with just a message.
// Code will default depending on the HTTP status code used in the response. Use respond.ErrorMessageWithCode() if you want to specify a custom code.
func ErrorMessage(message string) ErrorMessagePayload {
	return ErrorMessagePayload{Message: message}
}

// ErrorMessageWithCode is a helper function to create an ErrorMessagePayload with both message and code.
func ErrorMessageWithCode(message, code string) ErrorMessagePayload {
	return ErrorMessagePayload{Message: message, Code: code}
}

func createErrorResponse(httpCode int, payload ErrorMessagePayload) ErrorResponse {
	if payload.Message == "" {
		payload.Message = http.StatusText(httpCode)
	}

	return ErrorResponse{
		HttpCode: httpCode,
		Code:     payload.Code,
		Message:  payload.Message,
	}
}

func mergeErrorPayload(defaults ErrorMessagePayload, overrides ...ErrorMessagePayload) ErrorMessagePayload {
	if len(overrides) == 0 {
		return defaults
	}

	merged := defaults
	if overrides[0].Code != "" {
		merged.Code = overrides[0].Code
	}
	if overrides[0].Message != "" {
		merged.Message = overrides[0].Message
	}

	return merged
}

// Error issues an error response with the specified HTTP status code and message.
func Error(w http.ResponseWriter, e ErrorResponse) error {
	if e.Message == "" {
		e.Message = http.StatusText(e.HttpCode)
	}

	r := ErrorResponse{
		Code:    e.Code,
		Message: e.Message,
		Fields:  e.Fields,
	}

	return writeJson(w, e.HttpCode, r)
}

// ValidationError issues a validation error response with field-level details.
// It uses HTTP status code 422 (Unprocessable Entity) with code "validation_error" and a generic message.
func ValidationError(w http.ResponseWriter, fields map[string]string) error {
	r := ErrorResponse{
		HttpCode: http.StatusUnprocessableEntity,
		Code:     "validation_error",
		Message:  "One or more fields have validation errors.",
		Fields:   fields,
	}

	return Error(w, r)
}

// TooManyRequests issues a 429 Too Many Requests error response with a standard message.
func TooManyRequests(w http.ResponseWriter) error {
	r := ErrorResponse{
		HttpCode: http.StatusTooManyRequests,
		Code:     "rate_limit_exceeded",
		Message:  "Too many requests. Please wait and retry.",
	}

	return Error(w, r)
}

// ServiceUnavailable issues a 503 Service Unavailable error response with a standard message.
func ServiceUnavailable(w http.ResponseWriter, message ...ErrorMessagePayload) error {
	payload := mergeErrorPayload(ErrorMessagePayload{
		Code:    "service_unavailable",
		Message: "The service is temporarily unavailable.",
	}, message...)

	r := createErrorResponse(http.StatusServiceUnavailable, payload)
	return Error(w, r)
}

// GatewayTimeout issues a 504 Gateway Timeout error response with a standard message.
func GatewayTimeout(w http.ResponseWriter, message ...ErrorMessagePayload) error {
	payload := mergeErrorPayload(ErrorMessagePayload{
		Code:    "gateway_timeout",
		Message: "The upstream service timed out.",
	}, message...)

	r := createErrorResponse(http.StatusGatewayTimeout, payload)
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
func BadRequest(w http.ResponseWriter, message ...ErrorMessagePayload) error {
	payload := mergeErrorPayload(ErrorMessagePayload{
		Code:    "bad_request",
		Message: "Bad request.",
	}, message...)

	r := createErrorResponse(http.StatusBadRequest, payload)
	return Error(w, r)
}

// UnprocessableEntity issues a 422 Unprocessable Entity error response with a custom message.
// This status code is used when the request is syntactically correct but semantically invalid (e.g., fails business rules).
func UnprocessableEntity(w http.ResponseWriter, message ...ErrorMessagePayload) error {
	payload := mergeErrorPayload(ErrorMessagePayload{
		Code:    "unprocessable_entity",
		Message: "Unprocessable entity",
	}, message...)

	r := createErrorResponse(http.StatusUnprocessableEntity, payload)
	return Error(w, r)
}

// Unauthorized issues a 401 Unauthorized error response.
// This status code is used when authentication is required and has failed or has not yet been provided.
func Unauthorized(w http.ResponseWriter, message ...ErrorMessagePayload) error {
	payload := mergeErrorPayload(ErrorMessagePayload{
		Code:    "unauthorized",
		Message: "You are not authorized to access this resource.",
	}, message...)

	r := createErrorResponse(http.StatusUnauthorized, payload)
	return Error(w, r)
}

// Forbidden issues a 403 Forbidden error response with a standard message.
// This status code is used when the user is authenticated but does not have permission to access the resource.
func Forbidden(w http.ResponseWriter, message ...ErrorMessagePayload) error {
	payload := mergeErrorPayload(ErrorMessagePayload{
		Code:    "forbidden",
		Message: "You do not have permission to access this resource.",
	}, message...)

	r := createErrorResponse(http.StatusForbidden, payload)
	return Error(w, r)
}

// Conflict issues a 409 Conflict error response with a custom message.
// This status code is used when the request could not be completed due to a conflict with the current state of the resource (e.g., duplicate entry).
func Conflict(w http.ResponseWriter, message ...ErrorMessagePayload) error {
	payload := mergeErrorPayload(ErrorMessagePayload{
		Code:    "conflict",
		Message: "Conflict occurred.",
	}, message...)

	r := createErrorResponse(http.StatusConflict, payload)
	return Error(w, r)
}

// NotImplemented issues a 501 Not Implemented error response with a standard message.
func NotImplemented(w http.ResponseWriter, message ...ErrorMessagePayload) error {
	payload := mergeErrorPayload(ErrorMessagePayload{
		Code:    "not_implemented",
		Message: "Not implemented.",
	}, message...)

	r := createErrorResponse(http.StatusNotImplemented, payload)
	return Error(w, r)
}

// PayloadTooLarge issues a 413 Payload Too Large error response with a custom message if provided, otherwise a default message.
func PayloadTooLarge(w http.ResponseWriter, message ...string) error {
	msg := "Payload too large."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	r := ErrorResponse{
		HttpCode: http.StatusRequestEntityTooLarge,
		Code:     "payload_too_large",
		Message:  msg,
	}

	return Error(w, r)
}
