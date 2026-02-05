package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/common-lib/requests"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCheckValidSubscriptionType(t *testing.T) {
	v := validator.New()
	require.NoError(t, requests.RegisterValidation(v, CheckValidSubscriptionType))

	type dto struct {
		Type string `validate:"validsubtype"`
	}

	require.NoError(t, v.Struct(dto{Type: "GENRE"}))
	require.NoError(t, v.Struct(dto{Type: "ARTIST"}))
	require.Error(t, v.Struct(dto{Type: "invalid"}))
	require.Error(t, v.Struct(dto{Type: ""}))
}

func TestCheckValidEntityID(t *testing.T) {
	v := validator.New()
	require.NoError(t, requests.RegisterValidation(v, CheckValidEntityID))

	type dto struct {
		EntityID string `validate:"validentityid"`
	}

	require.NoError(t, v.Struct(dto{EntityID: primitive.NewObjectID().Hex()}))
	require.Error(t, v.Struct(dto{EntityID: "not-a-valid-objectid"}))
}
