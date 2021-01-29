package core

import (
	"github.com/DoNewsCode/std/pkg/observability"
)

func ProvideObservability(c *C) {
	c.Provide(observability.ProvideJaegerLogAdapter)
	c.Provide(observability.ProvideHistogramMetrics)
	c.Provide(observability.ProvideOpentracing)
}
