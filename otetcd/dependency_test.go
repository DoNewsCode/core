package otetcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

func TestMain(m *testing.M) {
	if !envDefaultEtcdAddrsIsSet {
		fmt.Println("Set env ETCD_ADDR to run otetcd tests")
		os.Exit(0)
	}

	os.Exit(m.Run())
}

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
	out, cleanup := provideFactory(factoryIn{
		Conf: config.MapAdapter{"etcd": map[string]Option{
			"default": {
				Endpoints: envDefaultEtcdAddrs,
			},
			"alternative": {
				Endpoints: envDefaultEtcdAddrs,
			},
		}},
		Logger: log.NewNopLogger(),
		Tracer: nil,
	})
	alt, err := out.Factory.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	def, err := out.Factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	cleanup()
}

func Test_provideConfig(t *testing.T) {
	conf := provideConfig()
	_, err := yaml.Marshal(conf.Config)
	assert.NoError(t, err)
}
