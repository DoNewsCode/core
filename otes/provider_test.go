package otes

import (
	"fmt"
	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("ELASTICSEARCH_ADDR") == "" {
		fmt.Println("Set env ELASTICSEARCH_ADDR to run otes tests")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestNewEsFactory(t *testing.T) {
	esFactory, cleanup := provideEsFactory(in{
		Conf: config.MapAdapter{"es": map[string]Config{
			"default":     {URL: []string{os.Getenv("ELASTICSEARCH_ADDR")}},
			"alternative": {URL: []string{os.Getenv("ELASTICSEARCH_ADDR")}},
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

func TestNewEsFactoryWithOptions(t *testing.T) {
	var called bool
	esFactory, cleanup := provideEsFactory(in{
		Conf: config.MapAdapter{"es": map[string]Config{
			"default": {URL: []string{os.Getenv("ELASTICSEARCH_ADDR")}},
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
}

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}
