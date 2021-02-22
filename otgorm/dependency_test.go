package otgorm

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestProvideDBFactory(t *testing.T) {
	factory, cleanup := provideDBFactory(DatabaseIn{
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
