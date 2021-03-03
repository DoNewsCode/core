package otes

import (
	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	esConfig "github.com/olivere/elastic/v7/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEsFactory(t *testing.T) {
	esFactory, cleanup := provideEsFactory(in{
		Conf: config.MapAdapter{"es": map[string]esConfig.Config{
			"default":     {URL: "http://localhost:9200"},
			"alternative": {URL: "http://localhost:9200"},
		}},
		Logger: log.NewNopLogger(),
		Tracer: nil,
	})
	def, err := esFactory.Maker.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := esFactory.Maker.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	assert.NotNil(t, cleanup)
	cleanup()
}

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}
