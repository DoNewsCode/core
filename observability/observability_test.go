package observability

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/stretchr/testify/assert"
)

func TestProvideOpentracing(t *testing.T) {
	conf, _ := config.NewConfig(config.WithProviderLayer(rawbytes.Provider([]byte(sample)), yaml.Parser()))
	Out, cleanup, err := ProvideOpentracing(
		config.AppName("foo"),
		config.EnvTesting,
		ProvideJaegerLogAdapter(log.NewNopLogger()),
		conf,
	)
	assert.NoError(t, err)
	assert.NotNil(t, Out)
	cleanup()
}

func TestProvideHistogramMetrics(t *testing.T) {
	Out := ProvideHistogramMetrics(
		config.AppName("foo"),
		config.EnvTesting,
	)
	assert.NotNil(t, Out)
}

func TestProvideGORMMetrics(t *testing.T) {
	Out := ProvideGORMMetrics(
		config.AppName("foo"),
		config.EnvTesting,
	)
	assert.NotNil(t, Out)
}

func TestExportedConfigs(t *testing.T) {
	Conf := exportConfig()
	assert.NotEmpty(t, Conf.Config)
}
