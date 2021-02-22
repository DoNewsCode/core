package kitmw

import (
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
)

// MakeRetryMiddleware returns a middleware that retries the failed requests.
func MakeRetryMiddleware(max int, timeout time.Duration) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return lb.Retry(max, timeout, lb.NewRoundRobin(sd.FixedEndpointer{next}))
	}
}
