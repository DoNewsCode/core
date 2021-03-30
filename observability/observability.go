package observability

import (
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
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
	return di.Deps{provide, provideConfig}
}

// in is the injection argument of provide.
type in struct {
	dig.In

	Logger  log.Logger
	Conf    contract.ConfigAccessor
	AppName contract.AppName
	Env     contract.Env
}

// out is the result of provide
type out struct {
	dig.Out

	Tracer opentracing.Tracer
	Hist   metrics.Histogram
}

// provide provides the observability suite for the system. It contains a tracer and
// a histogram to measure all incoming request.
func provide(in in) (out, func(), error) {
	in.Logger = log.With(in.Logger, "tag", "observability")
	jlogger := ProvideJaegerLogAdapter(in.Logger)
	tracer, cleanup, err := ProvideOpentracing(in.AppName, in.Env, jlogger, in.Conf)
	hist := ProvideHistogramMetrics(in.AppName, in.Env)
	return out{
		Tracer: tracer,
		Hist:   hist,
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
    addr: ''
`

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

func provideConfig() configOut {

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
