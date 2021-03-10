package dtx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ep(ctx context.Context, req interface{}) (resp interface{}, err error) {
	return req, nil
}

type oncer func(ctx context.Context, key string) bool

func (o oncer) Once(ctx context.Context, key string) bool {
	return o(ctx, key)
}

type locker struct {
	Lockf   func(ctx context.Context, key string) bool
	Unlockf func(ctx context.Context, key string)
}

func (l locker) Lock(ctx context.Context, key string) bool {
	return l.Lockf(ctx, key)
}

func (l locker) Unlock(ctx context.Context, key string) {
	l.Unlockf(ctx, key)
}

type transactioner struct {
	attempt func(ctx context.Context, s string) bool
	cancel  func(ctx context.Context, s string) bool
}

func (t transactioner) MarkCancelledCheckAttempted(ctx context.Context, s string) bool {
	return t.attempt(ctx, s)
}

func (t transactioner) MarkAttemptedCheckCancelled(ctx context.Context, s string) bool {
	return t.cancel(ctx, s)
}

func TestMakeIdempotence(t *testing.T) {
	t.Run("without context", func(t *testing.T) {
		t.Parallel()
		var attempt = 0
		var s = func(ctx context.Context, key string) bool {
			if attempt == 0 {
				attempt++
				return false
			}
			return true
		}
		m := MakeIdempotence(oncer(s))
		f := m(ep)
		resp, err := f(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp)
	})

	t.Run("with context", func(t *testing.T) {
		t.Parallel()
		var attempt = 0
		var s = func(ctx context.Context, key string) bool {
			if attempt == 0 {
				attempt++
				return false
			}
			return true
		}
		m := MakeIdempotence(oncer(s))
		f := m(ep)
		ctx := context.WithValue(context.Background(), CorrelationID, "foobar")
		resp, err := f(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp)
		resp, err = f(ctx, 2)
		assert.Error(t, err)
		assert.Equal(t, nil, resp)
	})
}

func TestMakeLock(t *testing.T) {
	t.Run("no context", func(t *testing.T) {
		t.Parallel()
		var lock = locker{
			Lockf: func(ctx context.Context, key string) bool {
				return true
			},
			Unlockf: func(ctx context.Context, key string) {
			},
		}
		m := MakeLock(lock)
		f := m(ep)
		resp, err := f(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp)
	})

	t.Run("with context", func(t *testing.T) {
		t.Parallel()
		var lock = locker{
			Lockf: func(ctx context.Context, key string) bool {
				return true
			},
			Unlockf: func(ctx context.Context, key string) {
			},
		}
		m := MakeLock(lock)
		f := m(ep)
		ctx := context.WithValue(context.Background(), CorrelationID, "foobar")
		resp, err := f(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp)
	})

	t.Run("failed to grab lock", func(t *testing.T) {
		t.Parallel()
		var lock = locker{
			Lockf: func(ctx context.Context, key string) bool {
				return false
			},
			Unlockf: func(ctx context.Context, key string) {
			},
		}
		m := MakeLock(lock)
		f := m(ep)
		ctx := context.WithValue(context.Background(), CorrelationID, "foobar")

		resp, err := f(ctx, 2)
		assert.Error(t, err)
		assert.Equal(t, nil, resp)
	})
}

func TestMakeAttempt(t *testing.T) {
	t.Run("no context", func(t *testing.T) {
		t.Parallel()
		var tr = transactioner{
			attempt: func(ctx context.Context, key string) bool {
				return false
			},
			cancel: func(ctx context.Context, key string) bool {
				return false
			},
		}
		f := MakeAttempt(tr)(ep)
		g := MakeCancel(tr)(ep)
		resp, err := f(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp)
		resp, err = g(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp)
	})

	t.Run("with context, attempted", func(t *testing.T) {
		t.Parallel()
		ctx := context.WithValue(context.Background(), CorrelationID, "foobar")
		var tr = transactioner{
			attempt: func(ctx context.Context, key string) bool {
				return true
			},
			cancel: func(ctx context.Context, key string) bool {
				return false
			},
		}
		f := MakeAttempt(tr)(ep)
		g := MakeCancel(tr)(ep)

		resp, err := f(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp)

		resp, err = g(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp)
	})

	t.Run("with context, not attempted", func(t *testing.T) {
		t.Parallel()
		ctx := context.WithValue(context.Background(), CorrelationID, "foobar")
		var tr = transactioner{
			attempt: func(ctx context.Context, key string) bool {
				return false
			},
			cancel: func(ctx context.Context, key string) bool {
				return false
			},
		}
		f := MakeAttempt(tr)(ep)
		g := MakeCancel(tr)(ep)

		resp, err := f(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, resp)

		resp, err = g(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, nil, resp)
	})

	t.Run("with context, cancelled", func(t *testing.T) {
		t.Parallel()
		ctx := context.WithValue(context.Background(), CorrelationID, "foobar")
		var tr = transactioner{
			attempt: func(ctx context.Context, key string) bool {
				return false
			},
			cancel: func(ctx context.Context, key string) bool {
				return true
			},
		}
		f := MakeAttempt(tr)(ep)

		resp, err := f(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, nil, resp)
	})
}
