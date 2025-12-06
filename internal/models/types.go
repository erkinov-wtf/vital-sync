package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/lib/pq"
)

// JSONB stores raw JSON to allow both objects and arrays.
type JSONB []byte

// Value implements the driver.Valuer interface for database storage.
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil || len(j) == 0 {
		return []byte("[]"), nil
	}
	return j, nil
}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("JSONB: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Scan implements the sql.Scanner interface for database retrieval.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONB value")
	}
	*j = bytes
	return nil
}

// Unmarshal helps decode JSONB into the provided destination.
func (j JSONB) Unmarshal(dest interface{}) error {
	if len(j) == 0 {
		return nil
	}
	return json.Unmarshal(j, dest)
}

// NewJSONB marshals any value into JSONB.
func NewJSONB(value interface{}) (JSONB, error) {
	if value == nil {
		return nil, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return JSONB(data), nil
}

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return pq.Array([]string(s)).Value()
}

func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	var arr []string
	if err := pq.Array(&arr).Scan(value); err != nil {
		return err
	}
	*s = StringArray(arr)
	return nil
}

type TimeArray []time.Time

func (t TimeArray) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return json.Marshal(t)
}

func (t *TimeArray) Scan(value interface{}) error {
	if value == nil {
		*t = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}
