package kitmw

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/DoNewsCode/std/pkg/srverr"
)

func MakeErrorMarshallerMiddleware() endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func() {
				if er := recover(); er != nil {
					err = fmt.Errorf("panic: %s", er)
				}
			}()
			response, err = e(ctx, request)
			if err != nil {
				var serverError srverr.ServerError
				if !errors.As(err, &serverError) {
					serverError = srverr.UnknownErr(err)
				}
				// Brings kerr.SeverError to the uppermost level
				return response, serverError
			}

			return response, nil
		}
	}
}
