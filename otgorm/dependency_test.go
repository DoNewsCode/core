package otgorm

import (
	"os"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/events"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestProvideDBFactory(t *testing.T) {
	if os.Getenv("MYSQL_DSN") == "" {
		t.Skip("set MYSQL_DSN to run TestProvideDBFactory")
		return
	}
	gorms := map[string]any{
		"default": map[string]any{
			"database": "sqlite",
			"dsn":      ":memory:",
		},
		"alternative": map[string]any{
			"database": "mysql",
			"dsn":      os.Getenv("MYSQL_DSN"),
		},
	}

	for driverName := range gorms {
		for _, reloadable := range []bool{true, false} {
			t.Run(driverName, func(t *testing.T) {
				dispatcher := &events.Event[contract.ConfigUnmarshaler]{}
				out, cleanup, _ := provideDBFactory(&providersOption{reloadable: reloadable})(factoryIn{
					Conf:          config.MapAdapter{"gorm": gorms},
					Logger:        log.NewNopLogger(),
					Tracer:        nil,
					OnReloadEvent: dispatcher,
				})
				defer cleanup()
				db, err := out.Factory.Make(driverName)
				assert.NoError(t, err)
				assert.NotNil(t, db)
				assert.Equal(
					t,
					reloadable,
					dispatcher.ListenerCount() == 1,
					"unexpected dispatcher count %d when reload = %t",
					dispatcher.ListenerCount(),
					reloadable,
				)
			})
		}
	}
}

func TestGorm(t *testing.T) {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Invoke(func(
		d1 Maker,
		d2 *Factory,
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
