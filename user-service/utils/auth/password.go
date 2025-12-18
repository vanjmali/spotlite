package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	cost = 12
)

// HashPassword encrypts a plaintext password using bcrypt with the configured cost.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// CompareHashAndPassword verifies a login password against the stored bcrypt hash.
func CompareHashAndPassword(storedHashedPassword, passwordFromLogin string) error {
	return bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(passwordFromLogin))
}
