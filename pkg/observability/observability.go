package observability

import (
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/ghodss/yaml"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
)

// ObservabilityIn is the injection argument of Provide.
type ObservabilityIn struct {
	dig.In

	Logger  log.Logger
	Conf    contract.ConfigAccessor
	AppName contract.AppName
	Env     contract.Env
}

// ObservabilityOut is the result of Provide
type ObservabilityOut struct {
	dig.Out

	Tracer         opentracing.Tracer
	Hist           metrics.Histogram
	ExportedConfig []config.ExportedConfig `group:"config,flatten"`
}

// Provide provides the observability suite for the system. It contains a tracer and
// a histogram to measure all incoming request.
func Provide(in ObservabilityIn) (ObservabilityOut, func(), error) {
	in.Logger = log.With(in.Logger, "component", "observability")
	jlogger := ProvideJaegerLogAdapter(in.Logger)
	tracer, cleanup, err := ProvideOpentracing(in.AppName, in.Env, jlogger, in.Conf)
	hist := ProvideHistogramMetrics(in.AppName, in.Env)
	return ObservabilityOut{
		Tracer:         tracer,
		Hist:           hist,
		ExportedConfig: exportConfig(),
	}, cleanup, err
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

func exportConfig() []config.ExportedConfig {

	var conf map[string]interface{}
	_ = yaml.Unmarshal([]byte(sample), conf)
	return []config.ExportedConfig{
		{
			Owner:   "observability",
			Data:    conf,
			Comment: "The observability configuration",
		},
	}
}
