package observability_test

import (
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/observability"
	"github.com/go-kit/kit/metrics"
	"github.com/opentracing/opentracing-go"
)

func Example() {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(observability.Provide)
	c.Invoke(func(tracer opentracing.Tracer, metrics metrics.Histogram) {
		start := time.Now()
		span := tracer.StartSpan("test")
		time.Sleep(time.Second)
		span.Finish()
		metrics.With("module", "service", "method").Observe(time.Since(start).Seconds())
	})
}
