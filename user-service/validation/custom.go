package validation

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// CheckStrongPassword enforces length and complexity constraints on passwords.
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

// CheckValidUsername validates allowed characters and length for usernames.
func CheckValidUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	if !regexp.MustCompile(`^[a-zA-Z0-9._]{4,20}$`).MatchString(username) {
		return false
	}

	hasText := regexp.MustCompile(`[a-zA-Z]`).MatchString(username)
	hasDigit := regexp.MustCompile(`\d`).MatchString(username)

	return hasText || hasDigit
}

// GetErrorMsg maps validator tags to human-friendly error messages.
func GetErrorMsg(fe validator.FieldError) string {

	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "alpha":
		return fe.Field() + " must contain only letters"
	case "min":
		return fe.Field() + " is too short"
	case "validusername":
		return "Username must be 4-20 chars, contain letters and numbers, and only use dots/underscores"
	case "strongpassword":
		return "Password must be 10-20 chars, contain upper/lower case, a number, and a special character"
	case "email":
		return "Invalid email format"
	case "nefield":
		return fmt.Sprintf("%s must not match with %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("Validation failed on the '%s' tag", fe.Tag())
	}
}
