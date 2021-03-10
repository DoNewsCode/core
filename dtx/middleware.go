package dtx

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
)

// ErrNonIdempotent is returned when an endpoint is requested more than once with the same CorrelationID.
var ErrNonIdempotent = errors.New("rejected repeated request")

// ErrNoLock is returned when the endpoint fail to fetch the distributed lock under the same CorrelationID.
var ErrNoLock = errors.New("failed to grab lock")

// Oncer should return true if the key has been observed before.
type Oncer interface {
	Once(ctx context.Context, key string) bool
}

// MakeIdempotence returns a middleware that ensures the next endpoint can only be executed once per CorrelationID.
func MakeIdempotence(s Oncer) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			correlationID, ok := ctx.Value(CorrelationID).(string)
			if !ok {
				return e(ctx, request)
			}
			if s.Once(ctx, correlationID) {
				return nil, ErrNonIdempotent
			}
			return e(ctx, request)
		}
	}
}

// Locker is an interface for the distributed lock.
type Locker interface {
	// Lock should return true only when it successfully grabs the lock.
	Lock(ctx context.Context, key string) bool
	Unlock(ctx context.Context, key string)
}

// MakeLock returns a middleware that ensures the next endpoint is never concurrently accessed.
func MakeLock(l Locker) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			correlationID, ok := ctx.Value(CorrelationID).(string)
			if !ok {
				return e(ctx, request)
			}
			if l.Lock(ctx, correlationID) {
				defer l.Unlock(ctx, correlationID)
				return e(ctx, request)
			}
			return nil, ErrNoLock
		}
	}
}

// Sequencer is an interface that shields against the disordering of
// attempt and cancel in a transactional context.
type Sequencer interface {
	MarkCancelledCheckAttempted(context.Context, string) bool
	MarkAttemptedCheckCancelled(context.Context, string) bool
}

// MakeAttempt returns a middleware that wraps around an attempt endpoint. If the
// this segment of the distributed transaction is already cancelled, the next
// endpoint will never be executed.
//
// If the forward operation arrives later than the compensating operation due to
// network exceptions, the forward operation must be discarded. Otherwise,
// resource suspension occurs.
func MakeAttempt(s Sequencer) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			correlationID, ok := ctx.Value(CorrelationID).(string)
			if !ok {
				return e(ctx, request)
			}
			if s.MarkAttemptedCheckCancelled(ctx, correlationID) {
				return nil, nil
			}
			return e(ctx, request)
		}
	}
}

// MakeCancel returns a middleware that wraps around the cancellation endpoint.
// It guarantees if this segment of the distributed transaction is never attempted,
// the cancellation endpoint will not be executed.
//
// Transaction participants may receive the compensation order before performing
// normal operations due to network exceptions. In this case, null compensation
// is required.
func MakeCancel(s Sequencer) endpoint.Middleware {
	return func(e endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			correlationID, ok := ctx.Value(CorrelationID).(string)
			if !ok {
				return e(ctx, request)
			}
			if !s.MarkCancelledCheckAttempted(ctx, correlationID) {
				return nil, nil
			}
			return e(ctx, request)
		}
	}
}
