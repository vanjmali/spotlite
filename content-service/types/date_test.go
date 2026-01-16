package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vanjmali/spotlite/content/types"
	"go.mongodb.org/mongo-driver/bson"
)

func TestDateUnmarshalJSON(t *testing.T) {
	var d types.Date
	require.NoError(t, json.Unmarshal([]byte(`"2024-06-01"`), &d))
	require.Equal(t, time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), d.Time)

	var empty types.Date
	require.NoError(t, json.Unmarshal([]byte(`""`), &empty))
	require.True(t, (&empty).IsZero())

	var badType types.Date
	require.Error(t, json.Unmarshal([]byte(`123`), &badType))

	var badFormat types.Date
	require.Error(t, json.Unmarshal([]byte(`"06-01-2024"`), &badFormat))
}

func TestDateMarshalJSON(t *testing.T) {
	d := types.Date{Time: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)}
	out, err := json.Marshal(&d)
	require.NoError(t, err)
	require.Equal(t, `"2024-06-01"`, string(out))

	zero := types.Date{}
	out, err = json.Marshal(&zero)
	require.NoError(t, err)
	require.Equal(t, "null", string(out))
}

func TestDateMarshalBSONValue(t *testing.T) {
	d := types.Date{Time: time.Date(2024, 6, 1, 15, 30, 0, 0, time.UTC)}
	tpe, data, err := d.MarshalBSONValue()
	require.NoError(t, err)
	require.Equal(t, bson.TypeDateTime, tpe)

	var tm time.Time
	require.NoError(t, bson.UnmarshalValue(tpe, data, &tm))
	require.Equal(t, time.Date(2024, 6, 1, 15, 30, 0, 0, time.UTC), tm.UTC())

	zero := types.Date{}
	tpe, data, err = zero.MarshalBSONValue()
	require.NoError(t, err)
	require.Equal(t, bson.TypeNull, tpe)
	require.Nil(t, data)
}

func TestDateUnmarshalBSONValue(t *testing.T) {
	tm := time.Date(2024, 6, 1, 15, 30, 0, 0, time.UTC)
	tpe, data, err := bson.MarshalValue(tm)
	require.NoError(t, err)

	var d types.Date
	require.NoError(t, d.UnmarshalBSONValue(tpe, data))
	require.Equal(t, time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), d.Time)

	var zero types.Date
	require.NoError(t, zero.UnmarshalBSONValue(bson.TypeNull, nil))
	require.True(t, (&zero).IsZero())
}
