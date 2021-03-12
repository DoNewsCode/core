package kitmw

import (
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
)

// RetryOption is the parameter to config the retry middleware.
type RetryOption struct {
	Max     int
	Timeout time.Duration
}

// Retry returns a middleware that retries the failed requests.
func Retry(opt RetryOption) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return lb.Retry(opt.Max, opt.Timeout, lb.NewRoundRobin(sd.FixedEndpointer{next}))
	}
}
