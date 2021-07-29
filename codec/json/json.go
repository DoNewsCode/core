// Package json provides the json codec.
//
// The code and tests in this package is derived from
// https://github.com/go-kratos/kratos under MIT license
// https://github.com/go-kratos/kratos/blob/main/LICENSE
package json

import (
	"encoding/json"
	"reflect"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Codec is a Codec implementation with json.
type Codec struct {
	prefix           string
	indent           string
	marshalOptions   protojson.MarshalOptions
	unmarshalOptions protojson.UnmarshalOptions
}

// Option is the type of functional options to codec
type Option func(*Codec)

// NewCodec creates a new json codec
func NewCodec(opts ...Option) Codec {
	var (
		codec          Codec
		marshalOptions = protojson.MarshalOptions{
			EmitUnpopulated: true,
		}
		unmarshalOptions = protojson.UnmarshalOptions{
			DiscardUnknown: true,
		}
	)
	codec.marshalOptions = marshalOptions
	codec.unmarshalOptions = unmarshalOptions
	for _, f := range opts {
		f(&codec)
	}
	return codec
}

// WithIndent allows the codec to indent json output while marshalling. It is
// useful when the json output is meant for humans to read.
func WithIndent(indent string) Option {
	return func(codec *Codec) {
		codec.indent = indent
		codec.marshalOptions.Multiline = true
		codec.marshalOptions.Indent = indent
	}
}

// Marshal serialize the interface{} to []byte
func (c Codec) Marshal(v interface{}) ([]byte, error) {
	if m, ok := v.(proto.Message); ok {
		return c.marshalOptions.Marshal(m)
	}
	if c.indent != "" {
		return json.MarshalIndent(v, c.prefix, c.indent)
	}
	return json.Marshal(v)
}

// Unmarshal deserialize the []byte to interface{}
func (c Codec) Unmarshal(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		rv = rv.Elem()
	}
	if m, ok := v.(proto.Message); ok {
		return c.unmarshalOptions.Unmarshal(data, m)
	} else if m, ok := reflect.Indirect(reflect.ValueOf(v)).Interface().(proto.Message); ok {
		return c.unmarshalOptions.Unmarshal(data, m)
	}
	return json.Unmarshal(data, v)
}
