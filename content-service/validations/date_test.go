package validations_test

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"github.com/vanjmali/spotlite/content/types"
	"github.com/vanjmali/spotlite/content/validations"
)

func TestCheckValidDateOnly(t *testing.T) {
	v := validator.New()
	require.NoError(t, requests.RegisterValidation(v, validations.CheckValidDateOnly))

	type dto struct {
		ReleaseDate types.Date `validate:"notzerodate"`
	}

	require.NoError(t, v.Struct(dto{ReleaseDate: types.Date{Time: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)}}))
	require.Error(t, v.Struct(dto{ReleaseDate: types.Date{}}))
}
