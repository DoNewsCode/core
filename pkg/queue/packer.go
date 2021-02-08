package queue

import (
	"bytes"
	"encoding/gob"
	"reflect"
)

type packer struct {
}

func (p packer) Compress(message interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(message); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p packer) Decompress(data []byte, message interface{}) error {
	buf := bytes.NewBuffer(data)
	if rvalue, ok := message.(reflect.Value); ok {
		return gob.NewDecoder(buf).DecodeValue(rvalue)
	}
	return gob.NewDecoder(buf).Decode(message)
}
