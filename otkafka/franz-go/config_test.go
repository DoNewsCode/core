package franz_go

import (
	"github.com/DoNewsCode/core/config"
	"github.com/stretchr/testify/assert"
	"testing"
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

	err := kf.Unmarshal("kafka.default", &conf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{"foo"}, conf.SeedBrokers)
}
