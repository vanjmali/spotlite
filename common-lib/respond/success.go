package respond

import (
	"net/http"
)

func writeMessageJson(w http.ResponseWriter, statusCode int, message string) error {
	r := map[string]string{"message": message}
	return writeJson(w, statusCode, r)
}

// OkJson sends a 200 OK response with the provided data as JSON.
func OkJson(w http.ResponseWriter, data any) error {
	return writeJson(w, http.StatusOK, data)
}

// Ok sends a 200 OK response with a message as JSON.
func Ok(w http.ResponseWriter, message string) error {
	return writeMessageJson(w, http.StatusOK, message)
}

// CreatedJson sends a 201 CreatedJson response with the provided data as JSON.
func CreatedJson(w http.ResponseWriter, data any) error {
	return writeJson(w, http.StatusCreated, data)
}

// Created sends a 201 Created response with a message as JSON.
func Created(w http.ResponseWriter, message string) error {
	return writeMessageJson(w, http.StatusCreated, message)
}

// NoContent sends a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
