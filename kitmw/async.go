package kitmw

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"
)

// MakeAsyncMiddleware returns a go kit middleware that calls the next handler in
// a detached goroutine. Timeout and cancellation of the previous context no
// logger apply to the detached goroutine, the tracing context however is
// carried over.
func MakeAsyncMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			span := opentracing.SpanFromContext(ctx)
			go func() {
				ctx := opentracing.ContextWithSpan(context.Background(), span)
				_, err = next(ctx, request)
				if err != nil {
					level.Warn(logger).Log("err", err.Error())
				}
			}()
			return nil, nil
		}
	}
}
