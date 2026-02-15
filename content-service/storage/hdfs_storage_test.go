package storage

import "testing"

func TestNormalizeDataTransferProtection(t *testing.T) {
	t.Run("accepts authentication", func(t *testing.T) {
		got, err := normalizeDataTransferProtection("authentication")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "authentication" {
			t.Fatalf("expected authentication, got %q", got)
		}
	})

	t.Run("accepts integrity", func(t *testing.T) {
		got, err := normalizeDataTransferProtection("integrity")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "integrity" {
			t.Fatalf("expected integrity, got %q", got)
		}
	})

	t.Run("normalizes privacy case and whitespace", func(t *testing.T) {
		got, err := normalizeDataTransferProtection("  Privacy ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "privacy" {
			t.Fatalf("expected privacy, got %q", got)
		}
	})

	t.Run("rejects unsupported value", func(t *testing.T) {
		_, err := normalizeDataTransferProtection("off")
		if err == nil {
			t.Fatal("expected error for unsupported value")
		}
	})

	t.Run("accepts none and empty", func(t *testing.T) {
		got, err := normalizeDataTransferProtection("none")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "" {
			t.Fatalf("expected empty protection for none, got %q", got)
		}

		got, err = normalizeDataTransferProtection("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "" {
			t.Fatalf("expected empty protection for empty value, got %q", got)
		}
	})
}
