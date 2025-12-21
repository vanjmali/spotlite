package validation

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
)

var CheckStrongPassword = requests.CustomValidator{
	Tag: "strongpassword",
	Func: func(fl validator.FieldLevel) bool {
		password := fl.Field().String()

		if !regexp.MustCompile(`^\S{10,20}$`).MatchString(password) {
			return false
		}

		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		hasDigit := regexp.MustCompile(`\d`).MatchString(password)
		hasSymbol := regexp.MustCompile(`[!@#$%^&*.]`).MatchString(password)

		return hasUpper && hasLower && hasDigit && hasSymbol
	},
	ErrorMessage: func(fe validator.FieldError) string {
		return "Password must be 10-20 characters long, contain at least one uppercase letter, one lowercase letter, one digit, one special character, and have no spaces."
	},
}

var CheckValidUsername = requests.CustomValidator{
	Tag: "validusername",
	Func: func(fl validator.FieldLevel) bool {
		username := fl.Field().String()

		if !regexp.MustCompile(`^[a-zA-Z0-9._]{4,20}$`).MatchString(username) {
			return false
		}

		hasText := regexp.MustCompile(`[a-zA-Z]`).MatchString(username)
		hasDigit := regexp.MustCompile(`\d`).MatchString(username)

		return hasText || hasDigit
	},
	ErrorMessage: func(fe validator.FieldError) string {
		return fe.Field() + " must be 4-20 characters long, can contain letters, numbers, dots, and underscores, and must include at least one letter or number."
	},
}

var CheckValidName = requests.CustomValidator{
	Tag: "validname",
	Func: func(fl validator.FieldLevel) bool {
		name := fl.Field().String()

		// Allow letters (including accented Unicode letters like ć, č, š, đ, etc.) and spaces/hyphens
		if !regexp.MustCompile(`^[\p{L}\s\-']{2,20}$`).MatchString(name) {
			return false
		}

		return true
	},
	ErrorMessage: func(fe validator.FieldError) string {
		return fe.Field() + " must be 2-20 characters long and can contain letters, spaces, hyphens, and apostrophes."
	},
}
