package utils

import (
	"log"
	"os"
	"strconv"
	"time"
)

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("environment variable %s is required", key)
	}

	return value
}

func MustGetDurationEnv(key string, multiplier time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("environment variable %s is required", key)
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		log.Fatalf("%s must be a positive integer: %v", key, err)
	}

	return time.Duration(parsed) * multiplier
}
