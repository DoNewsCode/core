package kitmw

import "github.com/go-kit/kit/endpoint"

// LabeledMiddleware is a mutated endpoint.Middleware. It receives an additional
// method name from the caller.
type LabeledMiddleware func(string, endpoint.Endpoint) endpoint.Endpoint
