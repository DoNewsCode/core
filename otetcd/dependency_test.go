package otetcd

import (
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/events"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

func TestEtcd(t *testing.T) {
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
		d4 *clientv3.Client,
	) {
		a := assert.New(t)
		a.NotNil(d1)
		a.NotNil(d2)
		a.NotEmpty(d3.Cfg)
		a.NotNil(d4)
	})
}

func TestProvideFactory(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("Set env ETCD_ADDR to run TestProvideFactory")
		return
	}
	addrs := strings.Split(os.Getenv("ETCD_ADDR"), ",")
	for _, c := range []struct {
		name   string
		reload bool
	}{
		{"reload", true},
		{"no reload", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			var dispatcher = &events.SyncDispatcher{}
			out, cleanup := provideFactory(&providersOption{reloadable: c.reload})(factoryIn{
				Conf: config.MapAdapter{"etcd": map[string]Option{
					"default": {
						Endpoints: addrs,
					},
					"alternative": {
						Endpoints: addrs,
					},
				}},
				Logger:     log.NewNopLogger(),
				Tracer:     nil,
				Dispatcher: dispatcher,
			})
			alt, err := out.Factory.Make("alternative")
			assert.NoError(t, err)
			assert.NotNil(t, alt)
			def, err := out.Factory.Make("default")
			assert.NoError(t, err)
			assert.NotNil(t, def)
			assert.Equal(t, c.reload, dispatcher.ListenerCount(events.OnReload) == 1)
			cleanup()
		})

	}

}

func Test_provideConfig(t *testing.T) {
	conf := provideConfig()
	assert.NotNil(t, conf)
}
