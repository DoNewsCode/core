package kitmw

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/pkg/errors"
)

func MakeLoggingMiddleware(logger log.Logger, service, method string, printTrace bool) endpoint.Middleware {
	return func(endpoint endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			l := logging.WithContext(level.Debug(logger), ctx)
			response, err = endpoint(ctx, request)
			if err != nil {
				_ = l.Log("err", err.Error())
				if stacktracer, ok := err.(interface{ StackTrace() errors.StackTrace }); printTrace && ok {
					fmt.Printf("\n%+v\n\n", stacktracer.StackTrace())
				}
			}
			_ = l.Log(
				"service",
				service,
				"method",
				method,
				"request",
				fmt.Sprintf("%+v", request),
				"response",
				fmt.Sprintf("%+v", response),
			)
			return response, err
		}
	}
}

func MakeLabeledLoggingMiddleware(logger log.Logger, module string, printTrace bool) LabeledMiddleware {
	return func(method string, endpoint endpoint.Endpoint) endpoint.Endpoint {
		return MakeLoggingMiddleware(logger, module, method, printTrace)(endpoint)
	}
}
