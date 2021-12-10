package config

import (
	"io/ioutil"
	"testing"

	"github.com/DoNewsCode/core/codec/yaml"
	yaml2 "github.com/knadh/koanf/parsers/yaml"
	"github.com/stretchr/testify/assert"
)

func TestCodecParser_Marshal(t *testing.T) {
	parser := CodecParser{yaml.Codec{}}
	data, err := parser.Marshal(map[string]interface{}{"foo": "bar", "baz": 1})
	assert.NoError(t, err)
	expected, _ := yaml2.Parser().Marshal(map[string]interface{}{"foo": "bar", "baz": 1})
	assert.Equal(t, expected, data)
}

func TestCodecParser_Unmarshal(t *testing.T) {
	raw, _ := ioutil.ReadFile("./testdata/codec_parser.yaml")
	parser := CodecParser{yaml.Codec{}}
	data, err := parser.Unmarshal(raw)
	assert.NoError(t, err)
	expected, _ := yaml2.Parser().Unmarshal(raw)
	assert.Equal(t, expected, data)
}
