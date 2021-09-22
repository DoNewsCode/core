package otes

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/events"
	"github.com/go-kit/kit/log"
	"github.com/olivere/elastic/v7"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
)

type Populator struct{}

func (p Populator) Populate(target interface{}) error {
	g := dig.New()
	g.Provide(func() log.Logger {
		return log.NewNopLogger()
	})
	g.Provide(func() opentracing.Tracer {
		return opentracing.NoopTracer{}
	})
	return di.IntoPopulator(g).Populate(target)
}

func TestNewEsFactory(t *testing.T) {
	if os.Getenv("ELASTICSEARCH_ADDR") == "" {
		t.Skip("set env ELASTICSEARCH_ADDR to run TestNewEsFactory")
		return
	}
	addrs := strings.Split(os.Getenv("ELASTICSEARCH_ADDR"), ",")
	t.Run("normal construction", func(t *testing.T) {
		esFactory, cleanup := provideEsFactory(&providersOption{})(factoryIn{
			Conf: config.MapAdapter{"es": map[string]Config{
				"default":     {URL: addrs},
				"alternative": {URL: addrs},
			}},
			Logger:    log.NewNopLogger(),
			Populator: Populator{},
		})
		def, err := esFactory.Make("default")
		assert.NoError(t, err)
		assert.NotNil(t, def)
		alt, err := esFactory.Make("alternative")
		assert.NoError(t, err)
		assert.NotNil(t, alt)
		assert.NotNil(t, cleanup)
		cleanup()
	})
	t.Run("with options", func(t *testing.T) {
		var calledConstructor bool
		var calledConfig bool
		esFactory, cleanup := provideEsFactory(

			&providersOption{
				clientConstructor: func(args ClientArgs) (*elastic.Client, error) {
					calledConstructor = true
					return newClient(args)
				},
				interceptor: func(name string, opt *Config) {
					calledConfig = true
				},
			},
		)(
			factoryIn{
				Conf: config.MapAdapter{"es": map[string]Config{
					"default": {URL: addrs},
				}},
				Logger:    log.NewNopLogger(),
				Populator: Populator{},
			},
		)
		def, err := esFactory.Make("default")
		assert.NoError(t, err)
		assert.NotNil(t, def)
		assert.True(t, calledConstructor)
		assert.True(t, calledConfig)
		cleanup()
	})

	t.Run("should not connect to ES", func(t *testing.T) {
		esFactory, cleanup := provideEsFactory(&providersOption{})(factoryIn{
			Conf: config.MapAdapter{"es": map[string]Config{
				// elasticsearch server doesn't exist at this port
				"default": {URL: []string{"http://127.0.0.1:9999"}},
			}},
			Logger:    log.NewNopLogger(),
			Populator: Populator{},
		})
		def, err := esFactory.Make("default")
		assert.NoError(t, err)
		assert.NotNil(t, def)
		cleanup()
	})

	t.Run("should not reload if the providersOption forbids", func(t *testing.T) {
		dispatcher := &events.SyncDispatcher{}
		esFactory, cleanup := provideEsFactory(&providersOption{})(factoryIn{
			Conf: config.MapAdapter{"es": map[string]Config{
				// elasticsearch server doesn't exist at this port
				"default": {URL: []string{"http://127.0.0.1:9999"}},
			}},
			Logger:     log.NewNopLogger(),
			Populator:  Populator{},
			Dispatcher: dispatcher,
		})
		defer cleanup()

		def1, err := esFactory.Make("default")
		assert.NoError(t, err)
		dispatcher.Dispatch(context.Background(), events.OnReload, events.OnReloadPayload{})

		def2, err := esFactory.Make("default")
		assert.NoError(t, err)

		assert.Same(t, def1, def2)
	})

	t.Run("should reload if the providersOption allows", func(t *testing.T) {
		dispatcher := &events.SyncDispatcher{}
		esFactory, cleanup := provideEsFactory(&providersOption{reloadable: true})(factoryIn{
			Conf: config.MapAdapter{"es": map[string]Config{
				// elasticsearch server doesn't exist at this port
				"default": {URL: []string{"http://127.0.0.1:9999"}},
			}},
			Logger:     log.NewNopLogger(),
			Populator:  Populator{},
			Dispatcher: dispatcher,
		})
		defer cleanup()

		def1, err := esFactory.Make("default")
		assert.NoError(t, err)
		dispatcher.Dispatch(context.Background(), events.OnReload, events.OnReloadPayload{})

		def2, err := esFactory.Make("default")
		assert.NoError(t, err)

		assert.NotSame(t, def1, def2)
	})
}

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}
