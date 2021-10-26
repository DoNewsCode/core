package clihttp

import (
	"github.com/DoNewsCode/core/di"
	"github.com/opentracing/opentracing-go"
)

/* Providers provides the http client as a dependency.
Depends On:
	opentracing.Tracer
Provides:
	*clihttp.Client
*/
func Providers(options ...Option) di.Deps {
	return di.Deps{
		func(tracer opentracing.Tracer) *Client {
			return NewClient(tracer, options...)
		},
	}
}
