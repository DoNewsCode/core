package otfranz

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/stretchr/testify/assert"
)

func Test_fromConfig(t *testing.T) {
	conf := Config{}
	opts := fromConfig(conf)
	assert.Len(t, opts, 3)

	kf := config.MapAdapter{"kafka": map[string]Config{
		"default": {
			SeedBrokers: []string{"foo"},
		},
	}}

	// There are many options that can not be decoded.
	// This test is necessary to prevent missing tags of "-".
	err := kf.Unmarshal("kafka.default", &conf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{"foo"}, conf.SeedBrokers)
}
