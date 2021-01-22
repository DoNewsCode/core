package kitmw

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"time"
)

func MakeLabeledMetricsMiddleware(his metrics.Histogram, service string) LabeledMiddleware {
	return func(name string, e endpoint.Endpoint) endpoint.Endpoint {
		return MakeMetricsMiddleware(his, service, name)(e)
	}
}

func MakeMetricsMiddleware(his metrics.Histogram, service, method string) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				his.With("service", service, "method", method).Observe(time.Since(begin).Seconds())
			}(time.Now())
			return e(ctx, request)
		}
	}
}
