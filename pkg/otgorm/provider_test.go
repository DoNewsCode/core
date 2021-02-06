package otgorm

import (
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
	"testing"
)

func TestProvideDBFactory(t *testing.T) {
	factory, cleanup := ProvideDBFactory(DatabaseParams{
		In: dig.In{},
		Conf: config.MapAdapter{"gorm": map[string]databaseConf{
			"default": {
				Database: "sqlite",
				Dsn:      "",
			},
			"alternative": {
				Database: "sqlite",
				Dsn:      "",
			},
		}},
		Logger: log.NewNopLogger(),
		Tracer: nil,
	})
	alt, err := factory.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	def, err := factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	cleanup()
}
