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

// Marshal serialize the interface{} to []byte
func (Codec) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// Unmarshal deserialize the []byte to interface{}
func (Codec) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
