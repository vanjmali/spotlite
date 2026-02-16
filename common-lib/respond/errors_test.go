package respond

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPayloadTooLarge_DefaultAndCustomMessage(t *testing.T) {
	t.Run("default message", func(t *testing.T) {
		rec := httptest.NewRecorder()

		if err := PayloadTooLarge(rec); err != nil {
			t.Fatalf("PayloadTooLarge returned error: %v", err)
		}

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("expected status 413, got %d", rec.Code)
		}

		var body ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if body.Code != "payload_too_large" {
			t.Fatalf("expected code payload_too_large, got %q", body.Code)
		}
		if body.Message != "Payload too large." {
			t.Fatalf("expected default message, got %q", body.Message)
		}
	})

	t.Run("custom message", func(t *testing.T) {
		rec := httptest.NewRecorder()

		if err := PayloadTooLarge(rec, "max 100MB"); err != nil {
			t.Fatalf("PayloadTooLarge returned error: %v", err)
		}

		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("expected status 413, got %d", rec.Code)
		}

		var body ErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if body.Message != "max 100MB" {
			t.Fatalf("expected custom message, got %q", body.Message)
		}
	})
}
