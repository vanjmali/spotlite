package entities

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
