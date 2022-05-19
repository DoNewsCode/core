package config

import (
	"encoding/json"
	"errors"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a type that describe a time duration. It is suitable for use in
// configurations as it implements a variety of serialization methods.
type Duration struct {
	time.Duration
}

// IsZero returns true if the Duration is the zero value.
func (d Duration) IsZero() bool {
	return d.Duration == 0
}

// MarshalYAML implements yaml.Marshaler
func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == "!!float" {
		return d.UnmarshalJSON([]byte(value.Value))
	}
	return d.UnmarshalJSON([]byte("\"" + value.Value + "\""))
}

// MarshalJSON implements json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

// MarshalText implements encoding.TextMarshaler
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
