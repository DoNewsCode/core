package observability

import (
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"gopkg.in/yaml.v3"
)

/*
Providers returns a set of providers available in package observability

	Depends On:
		log.Logger
		contract.ConfigAccessor
		contract.AppName
		contract.Env
	Provides:
		opentracing.Tracer
		metrics.Histogram
*/
func Providers() di.Deps {
	return di.Deps{
		ProvideJaegerLogAdapter,
		ProvideOpentracing,
		ProvideHistogramMetrics,
		ProvideGORMMetrics,
		exportConfig,
	}
}

const sample = `
jaeger:
  sampler:
    type: 'const'
    param: 1
  reporter:
    log:
      enable: false
    addr:
`

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

func exportConfig() configOut {

	var conf map[string]interface{}
	_ = yaml.Unmarshal([]byte(sample), &conf)
	configs := []config.ExportedConfig{
		{
			Owner:   "observability",
			Data:    conf,
			Comment: "The observability configuration",
		},
	}
	return configOut{Config: configs}
}
