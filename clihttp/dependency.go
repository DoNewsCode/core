package clihttp

import (
	"fmt"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/opentracing/opentracing-go"
)

/* Providers provides the http client as a dependency.
Depends On:
	opentracing.Tracer
Provides:
	*clihttp.Client
*/
func Providers(options ...ProvidersOptionFunc) di.Deps {
	var p providersOption
	for _, f := range options {
		f(&p)
	}
	return di.Deps{
		func(tracer opentracing.Tracer, populator contract.DIPopulator) (*Client, error) {
			if p.clientConstructor != nil {
				doer, err := p.clientConstructor(ClientArgs{populator})
				if err != nil {
					return nil, fmt.Errorf("contructing contract.HttpDoer: %w", err)
				}
				return NewClient(tracer, append(p.clientOptions, WithDoer(doer))...), nil
			}
			return NewClient(tracer, p.clientOptions...), nil
		},
		di.Bind(new(*Client), new(contract.HttpDoer)),
	}
}
