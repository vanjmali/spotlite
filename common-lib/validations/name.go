package validations

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
)

var regexValidName = regexp.MustCompile(`^[\p{L}\s\-']{2,20}$`)

var CheckValidName = requests.CustomValidator{
	Tag: "validname",
	Func: func(fl validator.FieldLevel) bool {
		name := fl.Field().String()

		// Allow letters (including accented Unicode letters like ć, č, š, đ, etc.) and spaces/hyphens
		if !regexValidName.MatchString(name) {
			return false
		}

		return true
	},
	ErrorMessage: func(fe validator.FieldError) string {
		return fe.Field() + " must be 2-20 characters long and can contain letters, spaces, hyphens, and apostrophes."
	},
}

