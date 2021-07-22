package config

import (
	"github.com/DoNewsCode/core/contract"
)

type CodecParser struct {
	Codec contract.Codec
}

func (c CodecParser) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	var m = make(map[string]interface{})
	if err := c.Codec.Unmarshal(bytes, &m); err != nil {
		return m, err
	}
	return m, nil
}

func (c CodecParser) Marshal(m map[string]interface{}) ([]byte, error) {
	return c.Codec.Marshal(m)
}
