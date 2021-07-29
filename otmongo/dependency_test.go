package otmongo

import (
	"fmt"
	"os"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
)

func TestMain(m *testing.M) {
	if !envDefaultMongoAddrIsSet {
		fmt.Println("Set env MONGO_ADDR to run otmongo tests")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestNewMongoFactory(t *testing.T) {
	t.Parallel()
	factory, cleanup := provideMongoFactory(factoryIn{
		In: dig.In{},
		Conf: config.MapAdapter{"mongo": map[string]struct{ Uri string }{
			"default": {
				Uri: envDefaultMongoAddr,
			},
			"alternative": {
				Uri: envDefaultMongoAddr,
			},
		}},
		Tracer: nil,
	})
	def, err := factory.Maker.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := factory.Maker.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	assert.NotNil(t, cleanup)
	cleanup()
}

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}
