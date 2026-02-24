package utils

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetEnv returns the environment variable value or a fallback when unset.
func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

// MustGetEnv fetches the environment variable or exits the program if it is missing.
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("FATAL: environment variable %s is required", key)
	}

	return value
}

// MustGetDurationEnv parses a positive integer env var and multiplies it by the provided duration.
func MustGetDurationEnv(key string, multiplier time.Duration) time.Duration {
	value := MustGetEnv(key)
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		log.Fatalf("FATAL: %s must be a positive integer: %v", key, err)
	}

	return time.Duration(parsed) * multiplier
}

// GetIntEnv parses an integer environment variable and returns fallback if unset or invalid.
func GetIntEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

// GetPositiveIntEnv parses an integer environment variable and returns fallback if unset, invalid or non-positive.
func GetPositiveIntEnv(key string, fallback int) int {
	parsed := GetIntEnv(key, fallback)
	if parsed <= 0 {
		return fallback
	}

	return parsed
}

// GetBoolEnv parses a boolean-like environment variable and returns fallback if unset or invalid.
func GetBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
