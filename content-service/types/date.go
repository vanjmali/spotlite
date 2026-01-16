package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// Date parses and formats dates in YYYY-MM-DD form.
type Date struct {
	time.Time
}

var dateFormat = "2006-01-02"

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return errors.New("release_date must be a string in YYYY-MM-DD format")
	}

	if s == "" {
		d.Time = time.Time{}
		return nil
	}

	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return errors.New("release_date must be in YYYY-MM-DD format")
	}

	d.Time = t.UTC()
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}

	return []byte(fmt.Sprintf("%q", d.Format(dateFormat))), nil
}

func (d *Date) IsZero() bool {
	return d.Time.IsZero()
}

// MarshalBSONValue stores Date as a native BSON datetime as UTC.
func (d *Date) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if d.Time.IsZero() {
		return bson.TypeNull, nil, nil
	}
	return bson.MarshalValue(d.UTC())
}

// UnmarshalBSONValue reads a BSON datetime and normalizes it to UTC.
func (d *Date) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bson.TypeNull {
		d.Time = time.Time{}
		return nil
	}

	var tm time.Time
	if err := bson.UnmarshalValue(t, data, &tm); err != nil {
		return err
	}

	d.Time = time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}
