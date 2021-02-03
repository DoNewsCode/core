package core

import (
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/observability"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
)

func ProvideObservability(c *C) {
	c.Provide(observability.ProvideJaegerLogAdapter)
	c.Provide(observability.ProvideHistogramMetrics)
	c.Provide(observability.ProvideOpentracing)
}

type ObservabilityIn struct {
	dig.In

	Logger  log.Logger
	Conf    contract.ConfigAccessor
	AppName contract.AppName
	Env     contract.Env
}

type ObservabilityOut struct {
	dig.Out

	Tracer opentracing.Tracer
	Hist   metrics.Histogram
}

func Observability(in ObservabilityIn) (ObservabilityOut, func(), error) {
	jlogger := observability.ProvideJaegerLogAdapter(in.Logger)
	tracer, cleanup, err := observability.ProvideOpentracing(in.AppName, in.Env, jlogger, in.Conf)
	hist := observability.ProvideHistogramMetrics(in.AppName, in.Env)
	return ObservabilityOut{
		Tracer: tracer,
		Hist:   hist,
	}, cleanup, err
}
