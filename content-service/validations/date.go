package validations

import (
	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/content/types"
)

var CheckValidDateOnly = requests.CustomValidator{
	Tag: "notzerodate",
	Func: func(fl validator.FieldLevel) bool {
		if v, ok := fl.Field().Interface().(types.Date); ok {
			return !v.IsZero()
		}
		if v, ok := fl.Field().Interface().(*types.Date); ok {
			return v != nil && !v.IsZero()
		}
		return false
	},
	ErrorMessage: func(fe validator.FieldError) string {
		return fe.Field() + " must be a valid date"
	},
}
