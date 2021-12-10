package config

import (
	"github.com/DoNewsCode/core/contract"
)

// CodecParser implements the Parser interface. It converts any contract.Codec to
// a valid config parser.
type CodecParser struct {
	Codec contract.Codec
}

// Unmarshal converts the bytes to map
func (c CodecParser) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	if err := c.Codec.Unmarshal(bytes, &m); err != nil {
		return m, err
	}
	return m, nil
}

// Marshal converts the map to bytes.
func (c CodecParser) Marshal(m map[string]interface{}) ([]byte, error) {
	return c.Codec.Marshal(m)
}
