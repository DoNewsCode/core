package kitmw

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
)

// Timeout returns a middleware that timeouts the request when the timer expired.
func Timeout(duration time.Duration) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx, cancel := context.WithTimeout(ctx, duration)
			defer cancel()
			return e(ctx, request)
		}
	}
}
