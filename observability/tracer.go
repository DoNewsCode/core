package observability

import (
	"fmt"
	"io"

	"github.com/DoNewsCode/core/contract"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegermetric "github.com/uber/jaeger-lib/metrics"
)

// ProvideOpentracing provides a opentracing.Tracer.
func ProvideOpentracing(
	appName contract.AppName,
	env contract.Env,
	log jaeger.Logger,
	conf contract.ConfigAccessor,
) (opentracing.Tracer, func(), error) {
	cfg := jaegercfg.Configuration{
		ServiceName: fmt.Sprintf("%s.%s", appName, env),
		Sampler: &jaegercfg.SamplerConfig{
			Type:  conf.String("jaeger.sampler.type"),
			Param: conf.Float64("jaeger.sampler.param"),
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           conf.Bool("jaeger.reporter.log"),
			LocalAgentHostPort: conf.String("jaeger.reporter.addr"),
		},
	}
	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := log
	jMetricsFactory := jaegermetric.NullFactory

	// Initialize tracer with a logger and a metrics factory
	var (
		canceler io.Closer
		err      error
	)
	tracer, canceler, err := cfg.NewTracer(jaegercfg.Logger(jLogger), jaegercfg.Metrics(jMetricsFactory))
	if err != nil {
		log.Error(fmt.Sprintf("Could not initialize jaeger tracer: %s", err.Error()))
		return nil, nil, err
	}
	closer := func() {
		if err := canceler.Close(); err != nil {
			log.Error(err.Error())
		}
	}

	return tracer, closer, nil
}
