package otetcd

import (
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/ghodss/yaml"
	"github.com/go-kit/kit/log"
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
	out, cleanup := provideFactory(factoryIn{
		Conf: config.MapAdapter{"etcd": map[string]Option{
			"default": {
				Endpoints: []string{"localhost:2379"},
			},
			"alternative": {
				Endpoints: []string{"localhost:2379"},
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

func TestExportedConfigs(t *testing.T) {
	conf := provideExportedConfigs()
	_, err := yaml.Marshal(conf)
	assert.NoError(t, err)
}
