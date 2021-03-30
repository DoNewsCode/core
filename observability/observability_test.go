package observability

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/stretchr/testify/assert"
)

func TestProvide(t *testing.T) {
	conf, _ := config.NewConfig(config.WithProviderLayer(rawbytes.Provider([]byte(sample)), yaml.Parser()))
	Out, cleanup, err := provide(in{
		Conf:    conf,
		Logger:  log.NewNopLogger(),
		AppName: config.AppName("foo"),
		Env:     config.EnvTesting,
	})
	assert.NoError(t, err)
	assert.NotNil(t, Out.Tracer)
	assert.NotNil(t, Out.Hist)
	cleanup()
}

func Test_provideConfig(t *testing.T) {
	Conf := provideConfig()
	assert.NotEmpty(t, Conf.Config)
}
