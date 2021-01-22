package kitmw

import "github.com/go-kit/kit/endpoint"

type LabeledMiddleware func(string, endpoint.Endpoint) endpoint.Endpoint
