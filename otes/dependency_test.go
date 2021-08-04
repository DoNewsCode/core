package otes

import (
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
)

func TestNewEsFactory(t *testing.T) {
	if os.Getenv("ELASTICSEARCH_ADDR") == "" {
		t.Skip("set env ELASTICSEARCH_ADDR to run TestNewEsFactory")
		return
	}
	addrs := strings.Split(os.Getenv("ELASTICSEARCH_ADDR"), ",")
	t.Run("normal construction", func(t *testing.T) {
		esFactory, cleanup := provideEsFactory(factoryIn{
			Conf: config.MapAdapter{"es": map[string]Config{
				"default":     {URL: addrs},
				"alternative": {URL: addrs},
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
	})
	t.Run("with options", func(t *testing.T) {
		var called bool
		esFactory, cleanup := provideEsFactory(factoryIn{
			Conf: config.MapAdapter{"es": map[string]Config{
				"default": {URL: addrs},
			}},
			Logger: log.NewNopLogger(),
			Options: []elastic.ClientOptionFunc{
				func(client *elastic.Client) error {
					called = true
					return nil
				},
			},
			Tracer: nil,
		})
		def, err := esFactory.Maker.Make("default")
		assert.NoError(t, err)
		assert.NotNil(t, def)
		assert.True(t, called)
		cleanup()
	})

	t.Run("should not connect to ES", func(t *testing.T) {
		esFactory, cleanup := provideEsFactory(factoryIn{
			Conf: config.MapAdapter{"es": map[string]Config{
				// elasticsearch server doesn't exist at this port
				"default": {URL: []string{"http://127.0.0.1:9999"}},
			}},
			Logger: log.NewNopLogger(),
			Tracer: nil,
		})
		def, err := esFactory.Maker.Make("default")
		assert.NoError(t, err)
		assert.NotNil(t, def)
		cleanup()
	})
}

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}
