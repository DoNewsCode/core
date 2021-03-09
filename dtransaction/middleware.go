package dtransaction

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
)

type Oncer interface {
	Once(ctx context.Context, key string) bool
}

func MakeIdempotence(s Oncer) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			correlationId, ok := ctx.Value(CorrelationId).(string)
			if !ok {
				return e(ctx, request)
			}
			if !s.Once(ctx, correlationId) {
				return nil, nil
			}
			return e(ctx, request)
		}
	}
}

type Locker interface {
	Lock(ctx context.Context, key string) bool
	Unlock(ctx context.Context, key string)
}

func MakeLock(l Locker) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			correlationId, ok := ctx.Value(CorrelationId).(string)
			if !ok {
				return e(ctx, request)
			}
			if l.Lock(ctx, correlationId) {
				defer l.Unlock(ctx, correlationId)
				return e(ctx, request)
			}
			return nil, errors.New("fails to grab lock")
		}
	}
}

type AtomicTransactioner interface {
	MarkCancelledCheckAttempted(context.Context, string) bool
	MarkAttemptedCheckCancelled(context.Context, string) bool
}

func MakeAttempt(s AtomicTransactioner) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			correlationId, ok := ctx.Value(CorrelationId).(string)
			if !ok {
				return e(ctx, request)
			}
			if s.MarkAttemptedCheckCancelled(ctx, correlationId) {
				return nil, nil
			}
			return e(ctx, request)
		}
	}
}

func MakeCancel(s AtomicTransactioner) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			correlationId, ok := ctx.Value(CorrelationId).(string)
			if !ok {
				return e(ctx, request)
			}
			if s.MarkCancelledCheckAttempted(ctx, correlationId) {
				return nil, nil
			}
			return e(ctx, request)
		}
	}
}
