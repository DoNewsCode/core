// Package yaml provides the yaml codec.
//
// The code and tests in this package is derived from
// https://github.com/go-kratos/kratos under MIT license
// https://github.com/go-kratos/kratos/blob/main/LICENSE
package yaml

import (
	"gopkg.in/yaml.v3"
)

// Codec is a Codec implementation with yaml.
type Codec struct{}

// Marshal serialize the v to []byte
func (Codec) Marshal(v any) ([]byte, error) {
	return yaml.Marshal(v)
}

// Unmarshal deserialize the []byte to v
func (Codec) Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}
