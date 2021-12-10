package observability_test

import (
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/observability"
	"github.com/DoNewsCode/core/srvhttp"
	"github.com/opentracing/opentracing-go"
)

func Example() {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(observability.Providers())
	c.Invoke(func(tracer opentracing.Tracer, metrics *srvhttp.RequestDurationSeconds) {
		start := time.Now()
		span := tracer.StartSpan("test")
		time.Sleep(time.Second)
		span.Finish()
		metrics.
			Module("module").
			Service("service").
			Route("route").
			Observe(time.Since(start).Seconds())
	})
}
