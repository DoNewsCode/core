package otgorm

import (
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
	"testing"
)

func TestNewDatabaseFactory(t *testing.T) {
	factory, cleanup, err := NewDatabaseFactory(DatabaseParams{
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
	assert.NoError(t, err)
	assert.Len(t, factory.db, 2)
	assert.NotNil(t, cleanup)
}
