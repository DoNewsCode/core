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
// logger apply to the detached goroutine, the tracing context however is carried
// over. A concurrency limit can be passed into the middleware. If the limit is
// reached, next endpoint call will block until the level of concurrency is below
// the limit.
func MakeAsyncMiddleware(logger log.Logger, concurrency int) endpoint.Middleware {
	limit := make(chan struct{}, concurrency)
	for i := 0; i < concurrency; i++ {
		limit <- struct{}{}
	}
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			span := opentracing.SpanFromContext(ctx)
			<-limit
			go func() {
				defer func() {
					limit <- struct{}{}
				}()

				var err error
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
