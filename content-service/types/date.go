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
	Value time.Time `bson:"-"` // Don't use default BSON marshaling, use our custom methods
}

var dateFormat = "2006-01-02"

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return errors.New("release_date must be a string in YYYY-MM-DD format")
	}

	if s == "" {
		d.Value = time.Time{}
		return nil
	}

	t, err := time.Parse(dateFormat, s)
	if err != nil {
		return errors.New("release_date must be in YYYY-MM-DD format")
	}

	d.Value = t.UTC()
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	if d.Value.IsZero() {
		return []byte("null"), nil
	}

	return fmt.Appendf(nil, "%q", d.Value.Format(dateFormat)), nil
}

func (d *Date) IsZero() bool {
	return d.Value.IsZero()
}

// Format returns the date formatted as YYYY-MM-DD
func (d *Date) Format(layout string) string {
	return d.Value.Format(layout)
}

// Time returns the underlying time.Time value
func (d *Date) Time() time.Time {
	return d.Value
}

// MarshalBSONValue stores Date as a BSON string in YYYY-MM-DD format.
func (d *Date) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if d.Value.IsZero() {
		return bson.TypeNull, nil, nil
	}
	dateStr := d.Value.Format(dateFormat)
	return bson.MarshalValue(dateStr)
}

// UnmarshalBSONValue reads a BSON value and normalizes it to UTC.
// It handles both BSON datetime and string formats.
func (d *Date) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bson.TypeNull {
		d.Value = time.Time{}
		return nil
	}

	// Handle BSON string type (for backward compatibility)
	if t == bson.TypeString {
		var s string
		if err := bson.UnmarshalValue(t, data, &s); err != nil {
			return err
		}
		return d.UnmarshalJSON([]byte(`"` + s + `"`))
	}

	// Handle BSON datetime type
	var tm time.Time
	if err := bson.UnmarshalValue(t, data, &tm); err != nil {
		return err
	}

	d.Value = time.Date(tm.Year(), tm.Month(), tm.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}
