package kitmw

import (
	"context"
	"errors"
	"fmt"

	"github.com/DoNewsCode/std/pkg/srverr"
	"github.com/go-kit/kit/endpoint"
)

type ErrorOption struct {
	AlwaysHTTP200 bool
	AlwaysGRPCOk  bool
	ShouldRecover bool
}

func MakeErrorMarshallerMiddleware(opt ErrorOption) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func() {
				if !opt.ShouldRecover {
					return
				}
				if er := recover(); er != nil {
					err = srverr.InternalErr(fmt.Errorf("panic: %s", er))
				}
			}()
			response, err = e(ctx, request)
			if err != nil {
				var serverError srverr.ServerError
				if !errors.As(err, &serverError) {
					serverError = srverr.UnknownErr(err)
				}
				if opt.AlwaysHTTP200 {
					serverError.HttpStatusCode = 200
				}
				if opt.AlwaysGRPCOk {
					serverError.GrpcStatusCode = 0
				}
				// Brings kerr.SeverError to the uppermost level
				return response, serverError
			}

			return response, nil
		}
	}
}
