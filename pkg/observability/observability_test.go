package observability

import (
	"testing"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/go-kit/kit/log"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/stretchr/testify/assert"
)

func TestProvide(t *testing.T) {
	conf, _ := config.NewConfig(config.WithProviderLayer(rawbytes.Provider([]byte(sample)), yaml.Parser()))
	Out, cleanup, err := Provide(ObservabilityIn{
		Conf:    conf,
		Logger:  log.NewNopLogger(),
		AppName: config.AppName("foo"),
		Env:     config.NewEnv("testing"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, Out.Tracer)
	assert.NotNil(t, Out.Hist)
	assert.NotNil(t, Out.ExportedConfig)
	cleanup()
}
