package otgorm

import (
	"os"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"gorm.io/gorm"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestProvideDBFactory(t *testing.T) {
	if os.Getenv("MYSQL_DSN") == "" {
		t.Skip("set MYSQL_DSN to run TestProvideDBFactory")
		return
	}
	gorms := map[string]databaseConf{
		"default": {
			Database: "sqlite",
			Dsn:      ":memory:",
		},
		"alternative": {
			Database: "mysql",
			Dsn:      os.Getenv("MYSQL_DSN"),
		},
	}
	out, cleanup, _ := provideDBFactory(factoryIn{
		Conf:   config.MapAdapter{"gorm": gorms},
		Logger: log.NewNopLogger(),
		Tracer: nil,
	})
	defer cleanup()
	for driverName := range gorms {
		t.Run(driverName, func(t *testing.T) {
			db, err := out.Maker.Make(driverName)
			assert.NoError(t, err)
			assert.NotNil(t, db)
		})
	}
}

func TestGorm(t *testing.T) {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Invoke(func(
		d1 Maker,
		d2 Factory,
		d3 struct {
			di.In
			Cfg []config.ExportedConfig `group:"config"`
		},
		d4 *gorm.DB,
	) {
		a := assert.New(t)
		a.NotNil(d1)
		a.NotNil(d2)
		a.NotEmpty(d3.Cfg)
		a.NotNil(d4)
	})
}

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}
