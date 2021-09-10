package ots3

import (
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/stretchr/testify/assert"
)

type Populator struct{}

func (p Populator) Populate(target interface{}) error { return nil }

func TestNewUploadManagerFactory(t *testing.T) {
	s3out := provideFactory(&providersOption{})(factoryIn{
		Conf: config.MapAdapter{"s3": map[string]S3Config{
			"default":     {},
			"alternative": {},
		}},
		Populator: Populator{},
	})
	def, err := s3out.Factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := s3out.Factory.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
}

func TestNewUploadManagerFactory_customOption(t *testing.T) {
	var called bool
	s3out := provideFactory(&providersOption{ctor: func(args ManagerArgs) (*Manager, error) {
		called = true
		return newManager(args)
	}})(factoryIn{
		Conf: config.MapAdapter{"s3": map[string]S3Config{
			"default":     {},
			"alternative": {},
		}},
		Populator: Populator{},
	})
	def, err := s3out.Factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.True(t, called)
}

type exportedConfig struct {
	di.In

	Conf []config.ExportedConfig `group:"config"`
}

func TestProvideConfigs(t *testing.T) {
	c := core.New()
	c.Provide(di.Deps{provideConfig})
	c.Invoke(func(e exportedConfig) {
		assert.Equal(t, provideConfig().Config, e.Conf)
	})
}
