package validations_test

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/common-lib/validations"
)

func TestCheckValidName(t *testing.T) {
	v := validator.New()
	require.NoError(t, requests.RegisterValidation(v, validations.CheckValidName))

	type dto struct {
		Name string `validate:"validname"`
	}

	require.NoError(t, v.Struct(dto{Name: "Jane Doe"}))
	require.NoError(t, v.Struct(dto{Name: "O'Neil"}))
	require.NoError(t, v.Struct(dto{Name: "Anne-Marie"}))
	require.Error(t, v.Struct(dto{Name: "A"}))
	require.Error(t, v.Struct(dto{Name: "John3"}))
}
