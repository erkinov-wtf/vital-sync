package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

// Value converts TimeArray to PostgreSQL array format
func (t TimeArray) Value() (driver.Value, error) {
	if t == nil || len(t) == 0 {
		return "{}", nil
	}

	// PostgreSQL array format: {"08:00:00","20:00:00"}
	var strs []string
	for _, v := range t {
		// Format as TIME (HH:MM:SS)
		strs = append(strs, fmt.Sprintf("\"%s\"", v.Format("15:04:05")))
	}

	return "{" + strings.Join(strs, ",") + "}", nil
}

// Scan converts PostgreSQL array to TimeArray
func (t *TimeArray) Scan(value interface{}) error {
	if value == nil {
		*t = TimeArray{}
		return nil
	}

	// Handle PostgreSQL array format
	switch v := value.(type) {
	case []byte:
		return t.parseArray(string(v))
	case string:
		return t.parseArray(v)
	default:
		return fmt.Errorf("cannot scan type %T into TimeArray", value)
	}
}

func (t *TimeArray) parseArray(s string) error {
	// Remove braces
	s = strings.Trim(s, "{}")
	if s == "" {
		*t = TimeArray{}
		return nil
	}

	parts := strings.Split(s, ",")
	times := make([]time.Time, 0, len(parts))

	for _, part := range parts {
		part = strings.Trim(part, "\"")
		parsedTime, err := time.Parse("15:04:05", part)
		if err != nil {
			return err
		}
		times = append(times, parsedTime)
	}

	*t = times
	return nil
}
