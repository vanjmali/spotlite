package validation

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

func CheckStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if !regexp.MustCompile(`^\S{10,20}$`).MatchString(password) {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`\d`).MatchString(password)
	hasSymbol := regexp.MustCompile(`[!@#$%^&*.]`).MatchString(password)

	return hasUpper && hasLower && hasDigit && hasSymbol
}

func CheckValidUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	if !regexp.MustCompile(`^[a-zA-Z0-9._]{4,20}$`).MatchString(username) {
		return false
	}

	hasText := regexp.MustCompile(`[a-zA-Z]`).MatchString(username)
	hasDigit := regexp.MustCompile(`\d`).MatchString(username)

	return hasText || hasDigit
}

func GetErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", fe.Field())
	case "min":
		return fmt.Sprintf("%s is too short", fe.Field())
	case "validusername":
		return "Username must be 4-20 chars, contain letters and numbers, and only use dots/underscores"
	case "strongpassword":
		return "Password must be 10-20 chars, contain upper/lower case, a number, and a special character"
	case "email":
		return "Invalid email format"
	default:
		return fmt.Sprintf("Validation failed on the '%s' tag", fe.Tag())
	}
}
