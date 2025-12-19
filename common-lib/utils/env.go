package utils

import (
	"log"
	"os"
	"strconv"
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
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("FATAL: environment variable %s is required", key)
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		log.Fatalf("FATAL: %s must be a positive integer: %v", key, err)
	}

	return time.Duration(parsed) * multiplier
}
