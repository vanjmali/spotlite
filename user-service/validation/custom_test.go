package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/common-lib/requests"
)

func TestCheckStrongPassword(t *testing.T) {
	v := validator.New()
	require.NoError(t, requests.RegisterValidation(v, CheckStrongPassword))

	type dto struct {
		Password string `validate:"strongpassword"`
	}

	require.NoError(t, v.Struct(dto{Password: "StrongPass123!"}))
	require.Error(t, v.Struct(dto{Password: "short1A!"}))
	require.Error(t, v.Struct(dto{Password: "NoDigitPass!"}))
	require.Error(t, v.Struct(dto{Password: "nouppercase1!"}))
	require.Error(t, v.Struct(dto{Password: "NOLOWERCASE1!"}))
	require.Error(t, v.Struct(dto{Password: "Has Space1!"}))
}

func TestCheckValidUsername(t *testing.T) {
	v := validator.New()
	require.NoError(t, requests.RegisterValidation(v, CheckValidUsername))

	type dto struct {
		Username string `validate:"validusername"`
	}

	require.NoError(t, v.Struct(dto{Username: "user.name_1"}))
	require.Error(t, v.Struct(dto{Username: "abc"}))
	require.Error(t, v.Struct(dto{Username: "invalid-username"}))
	require.Error(t, v.Struct(dto{Username: "...."}))
}

func TestCheckValidName(t *testing.T) {
	v := validator.New()
	require.NoError(t, requests.RegisterValidation(v, CheckValidName))

	type dto struct {
		Name string `validate:"validname"`
	}

	require.NoError(t, v.Struct(dto{Name: "Jane Doe"}))
	require.NoError(t, v.Struct(dto{Name: "O'Neil"}))
	require.NoError(t, v.Struct(dto{Name: "Anne-Marie"}))
	require.Error(t, v.Struct(dto{Name: "A"}))
	require.Error(t, v.Struct(dto{Name: "John3"}))
}
