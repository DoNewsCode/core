package sagas

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type SetOncer interface {
	SetOnce(key string) bool
}

func NewIdempotence(s SetOncer) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			correlationId, ok := ctx.Value(CorrelationId).(string)
			if !ok {
				return e(ctx, request)
			}
			if !s.SetOnce(correlationId) {
				return nil, nil
			}
			return e(ctx, request)
		}
	}
}
