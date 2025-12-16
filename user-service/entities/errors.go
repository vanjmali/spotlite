package entities

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
