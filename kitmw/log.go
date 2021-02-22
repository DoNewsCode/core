package kitmw

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/key"
	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

// MakeLoggingMiddleware returns a middleware the logs every request and response at debug
// level.
func MakeLoggingMiddleware(logger log.Logger, keyer contract.Keyer, printTrace bool) endpoint.Middleware {
	return func(endpoint endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			l := logging.WithContext(level.Debug(logger), ctx)
			l = log.With(logger, key.SpreadInterface(keyer)...)
			response, err = endpoint(ctx, request)
			if err != nil {
				_ = l.Log("err", err.Error())
				if stacktracer, ok := err.(interface{ StackTrace() errors.StackTrace }); printTrace && ok {
					fmt.Printf("\n%+v\n\n", stacktracer.StackTrace())
				}
			}
			_ = l.Log(
				"request", fmt.Sprintf("%+v", request),
				"response", fmt.Sprintf("%+v", response),
			)
			return response, err
		}
	}
}

func MakeLabeledLoggingMiddleware(logger log.Logger, keyer contract.Keyer, printTrace bool) LabeledMiddleware {
	return func(method string, endpoint endpoint.Endpoint) endpoint.Endpoint {
		return MakeLoggingMiddleware(logger, key.With(keyer, "method", method), printTrace)(endpoint)
	}
}
