package kitmw

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
)

func MakeTimeoutMiddleware(duration time.Duration) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()
			return e(ctx, request)
		}
	}
}
