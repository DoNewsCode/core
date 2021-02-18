package observability

import (
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
)

// ObservabilityIn is the injection argument of ProvideObservability.
type ObservabilityIn struct {
	dig.In

	Logger  log.Logger
	Conf    contract.ConfigAccessor
	AppName contract.AppName
	Env     contract.Env
}

// ObservabilityOut is the result of ProvideObservability
type ObservabilityOut struct {
	dig.Out

	Tracer opentracing.Tracer
	Hist   metrics.Histogram
}

// ProvideObservability provides the observability suite for the system. It contains a tracer and
// a histogram to measure all incoming request.
func ProvideObservability(in ObservabilityIn) (ObservabilityOut, func(), error) {
	in.Logger = log.With(in.Logger, "component", "observability")
	jlogger := ProvideJaegerLogAdapter(in.Logger)
	tracer, cleanup, err := ProvideOpentracing(in.AppName, in.Env, jlogger, in.Conf)
	hist := ProvideHistogramMetrics(in.AppName, in.Env)
	return ObservabilityOut{
		Tracer: tracer,
		Hist:   hist,
	}, cleanup, err
}
