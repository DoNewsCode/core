package kitmw

import (
	"context"

	"github.com/DoNewsCode/std/pkg/unierr"
	"github.com/go-kit/kit/endpoint"
)

type validator interface {
	Validate() error
}

// MakeValidationMiddleware returns a middleware that validates the request by
// calling Validate(). The request must implement the following interface,
// otherwise the middleware is a no-op:
//
//  type validator interface {
//	 Validate() error
//  }
//
func MakeValidationMiddleware() endpoint.Middleware {
	return func(in endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			if t, ok := req.(validator); ok {
				err = t.Validate()
				if err != nil {
					return nil, unierr.InvalidArgumentErr(err)
				}
			}
			resp, err = in(ctx, req)
			return
		}
	}
}
