package kitmw

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/DoNewsCode/core/srvhttp"
	"google.golang.org/grpc/codes"

	"github.com/DoNewsCode/core/unierr"
	"github.com/go-kit/kit/endpoint"
)

// ErrorOption is an option that tunes the middleware returned by
// Error
type ErrorOption struct {
	AlwaysHTTP200 bool
	ShouldRecover bool
}

// Error returns a middleware that wraps the returned
// error from next handler with a *unierr.Error. if a successful response is
// returned from the next handler, this is a no op. If the error returned by next
// handler is already a *unierr.Error, this decorates the *unierr.Error based on
// ErrorOption.
func Error(opt ErrorOption) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func() {
				if !opt.ShouldRecover {
					return
				}
				if er := recover(); er != nil {
					err = unierr.InternalErr(fmt.Errorf("panic: %s", er))
				}
			}()
			response, err = e(ctx, request)
			if err != nil {
				var serverError *unierr.Error
				if !errors.As(err, &serverError) {
					serverError = unierr.UnknownErr(err)
				}
				if opt.AlwaysHTTP200 {
					serverError.HttpStatusCodeFunc = func(code codes.Code) int {
						return 200
					}
				}
				// Brings kerr.SeverError to the uppermost level
				return response, serverError
			}

			return response, nil
		}
	}
}

// ErrorEncoder is a go kit style http error encoder. Internally it uses
// srvhttp.ResponseEncoder to encode the error.
func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	encoder := srvhttp.NewResponseEncoder(w)
	encoder.EncodeError(err)
}
