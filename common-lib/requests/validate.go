package requests

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/vanjmali/spotlite/common-lib/respond"
)

type ValidatorCustomMessage func(fe validator.FieldError) string

type CustomValidator struct {
	// Tag is the name of the validation tag.
	// Used within struct field tags to invoke this custom validation.
	Tag string
	// Func is the validation function.
	// It should return true if the field is valid, false otherwise.
	Func validator.Func
	// ErrorMessage is the custom error message to display when reporting validation failures.
	ErrorMessage ValidatorCustomMessage
}

var customValidationMessages map[string]ValidatorCustomMessage = make(map[string]ValidatorCustomMessage)

// RegisterValidationMessage registers a custom error message for a validation tag.
func RegisterValidationMessage(tag string, msg ValidatorCustomMessage) {
	customValidationMessages[tag] = msg
}

// RegisterValidation registers a custom validator with the given validator instance.
// It also registers the associated custom error message for that validator.
func RegisterValidation(v *validator.Validate, c CustomValidator) error {
	if c.ErrorMessage == nil {
		return errors.New("custom validator must have an error message function")
	}

	if err := v.RegisterValidation(c.Tag, c.Func); err != nil {
		return fmt.Errorf("failed to register validation '%s': %w", c.Tag, err)
	}

	return nil
}

// GetErrorMsg maps validator tags to human-friendly error messages.
//
// It returns a specific message for known tags, including some built-in
// and defined custom validations using RegisterValidation.
//
// For unknown tags, it returns the default error message.
func getErrorMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "alpha":
		return fe.Field() + " must contain only letters"
	case "min":
		return fe.Field() + " is too short"
	case "email":
		return "Invalid email format"
	default:
		if msgFunc, exists := customValidationMessages[fe.Tag()]; exists {
			return msgFunc(fe)
		}
		return fe.Error() // default error message
	}
}

// ReadAndValidateJson reads JSON data from the provided io.ReadCloser into the given DTO.
//
// ReadAndValidateJson also validates the DTO using the provided validator instance
// and handles common errors such as empty body, malformed JSON, and validation failures
// back to the http.ResponseWriter.
//
// It returns a boolean indicating success, and an error if one occurred during processing.
func ReadAndValidateJson(w http.ResponseWriter, v *validator.Validate, rBody io.ReadCloser, dto any) (bool, error) {
	defer rBody.Close()

	if err := json.NewDecoder(rBody).Decode(&dto); err != nil {
		if err == io.EOF {
			return false, respond.BadRequest(w, "Request body can't be empty.")
		}

		syntaxError := &json.SyntaxError{}
		if errors.As(err, &syntaxError) {
			return false, respond.BadRequest(w, "Invalid JSON format: "+err.Error()+".")
		}

		return false, respond.InternalServerError(w)
	}

	if err := v.Struct(dto); err != nil {
		invalidValidationError := &validator.InvalidValidationError{}
		if errors.As(err, &invalidValidationError) {
			return false, respond.InternalServerError(w)
		}

		fieldErrors := make(map[string]string)

		var validationErrs validator.ValidationErrors
		if ok := errors.As(err, &validationErrs); !ok {
			return false, respond.InternalServerError(w)
		}

		for _, verr := range validationErrs {
			fieldErrors[verr.Field()] = getErrorMsg(verr)
		}

		return false, respond.ValidationError(w, fieldErrors)
	}

	return true, nil
}

func RegisterCommonValidationMessages() {
	RegisterValidationMessage("nefield", func(fe validator.FieldError) string {
		return fmt.Sprintf("%s must not match with %s", fe.Field(), fe.Param())
	})
}
