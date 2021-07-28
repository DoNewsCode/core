package ots3

import (
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/stretchr/testify/assert"
)

func TestNewUploadManagerFactory(t *testing.T) {
	s3out := provideFactory(in{
		Conf: config.MapAdapter{"s3": map[string]S3Config{
			"default":     {},
			"alternative": {},
		}},
		Tracer: nil,
	})
	def, err := s3out.Factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := s3out.Factory.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
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
