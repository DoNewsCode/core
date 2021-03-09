// +build integration

package dtransaction

import (
	"context"
	"testing"
	"time"

	"github.com/DoNewsCode/core/key"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestOnce(t *testing.T) {
	s := RedisStore{
		keyer:  key.New(),
		client: redis.NewUniversalClient(&redis.UniversalOptions{}),
	}
	ctx := context.Background()
	defer s.client.Del(ctx, "once:foobar")

	assert.False(t, s.Once(ctx, "foobar"))
	assert.True(t, s.Once(ctx, "foobar"))
	assert.True(t, s.Once(ctx, "foobar"))

}

func TestLock(t *testing.T) {
	s := RedisStore{
		keyer:  key.New(),
		client: redis.NewUniversalClient(&redis.UniversalOptions{}),
	}
	ctx := context.Background()
	defer s.client.Del(ctx, "lock:foobar")

	assert.True(t, s.Lock(ctx, "foobar"))

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond)
	defer cancel()
	assert.False(t, s.Lock(ctx, "foobar"))

	s.Unlock(ctx, "foobar")
}

func TestRedisStore_MarkAttemptedCheckCancelled(t *testing.T) {
	s := RedisStore{
		keyer:  key.New(),
		client: redis.NewUniversalClient(&redis.UniversalOptions{}),
	}
	ctx := context.Background()
	defer s.client.Del(ctx, "attempt:foobar")
	defer s.client.Del(ctx, "cancel:foobar")

	assert.False(t, s.MarkCancelledCheckAttempted(ctx, "foobar"))
	assert.True(t, s.MarkAttemptedCheckCancelled(ctx, "foobar"))

	s.client.Del(ctx, "attempt:foobar")
	s.client.Del(ctx, "cancel:foobar")

	assert.False(t, s.MarkAttemptedCheckCancelled(ctx, "foobar"))
	assert.True(t, s.MarkCancelledCheckAttempted(ctx, "foobar"))
}
