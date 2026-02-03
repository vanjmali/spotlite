package validation

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/subscription"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var CheckValidSubscriptionType = requests.CustomValidator{
	Tag: "validsubtype",
	Func: func(fl validator.FieldLevel) bool {
		subType := fl.Field().String()

		if !subscription.IsValidSubscriptionType(subscription.SubscriptionType(subType)) {
			return false
		}

		return true
	},
	ErrorMessage: func(fe validator.FieldError) string {
		return "subscription type must be valid."
	},
}

var CheckValidEntityID = requests.CustomValidator{
	Tag: "validentityid",
	Func: func(fl validator.FieldLevel) bool {
		field := fl.Field()

		if field.Kind() != reflect.String {
			return false
		}

		_, err := primitive.ObjectIDFromHex(field.String())
		return err == nil
	},
	ErrorMessage: func(fe validator.FieldError) string {
		return "entity ID must be valid."
	},
}
