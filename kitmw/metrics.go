package kitmw

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/key"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
)

// MakeLabeledMetricsMiddleware returns a LabeledMiddleware that collects histogram metrics.
func MakeLabeledMetricsMiddleware(his metrics.Histogram, keyer contract.Keyer) LabeledMiddleware {
	return func(name string, e endpoint.Endpoint) endpoint.Endpoint {
		return MakeMetricsMiddleware(his, key.With(keyer, "method", name))(e)
	}
}

// MakeMetricsMiddleware returns a middleware that collects histogram metrics.
func MakeMetricsMiddleware(his metrics.Histogram, keyer contract.Keyer) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				his.With(keyer.Spread()...).Observe(time.Since(begin).Seconds())
			}(time.Now())
			return e(ctx, request)
		}
	}
}
