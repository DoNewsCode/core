package ctxmeta

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithoutCancel(t *testing.T) {
	key := struct{}{}
	ctx := context.WithValue(context.Background(), key, "value")
	ctx, cancel := context.WithCancel(ctx)
	cancel()

	select {
	case <-WithoutCancel(ctx).Done():
		t.Fatal("context is cancelled")
	default:
	}

	_, dead := WithoutCancel(ctx).Deadline()
	assert.False(t, dead)
	assert.Nil(t, WithoutCancel(ctx).Err())
	assert.Equal(t, "value", WithoutCancel(ctx).Value(key))
}

func TestWithoutCancel_Nil(t *testing.T) {
	defer func() {
		assert.Equal(t, recover(), "cannot create context from nil parent")
	}()
	WithoutCancel(nil)
}
