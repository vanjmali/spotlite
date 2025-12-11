package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	cost := 12

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
