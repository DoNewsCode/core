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

// MarshalYAML implements Marshaller
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.String(), nil
}

// UnmarshalYAML implements Unmarshaller
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == "!!float" {
		return d.UnmarshalJSON([]byte(value.Value))
	}
	return d.UnmarshalJSON([]byte("\"" + value.Value + "\""))
}

// MarshalJSON implements JSONMarshaller
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON implements JSONUnmarsheller
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
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
