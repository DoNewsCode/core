package otmongo

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/events"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
)

func TestNewMongoFactory(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name       string
		reloadable bool
	}{
		{
			"reload", true,
		},
		{
			"not reload", false,
		},
	} {
		dispatcher := &events.SyncDispatcher{}
		factory, cleanup := provideMongoFactory(&providersOption{
			reloadable: c.reloadable,
		})(factoryIn{
			In: dig.In{},
			Conf: config.MapAdapter{"mongo": map[string]struct {
				URI string `json:"uri"`
			}{
				"default": {
					URI: "mongodb://127.0.0.1:27017",
				},
				"alternative": {
					URI: "mongodb://127.0.0.1:27017",
				},
			}},
			Tracer:     nil,
			Dispatcher: dispatcher,
		})
		def, err := factory.Make("default")
		assert.NoError(t, err)
		assert.NotNil(t, def)
		alt, err := factory.Make("alternative")
		assert.NoError(t, err)
		assert.NotNil(t, alt)
		assert.NotNil(t, cleanup)
		assert.Equal(t, c.reloadable, dispatcher.ListenerCount(events.OnReload) == 1)
		cleanup()
	}
}

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}
