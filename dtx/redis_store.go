package dtx

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/go-redis/redis/v8"
)

// RedisStore is an implementation of Oncer, Locker and Sequencer.
type RedisStore struct {
	keyer  contract.Keyer
	client redis.UniversalClient
}

// MarkCancelledCheckAttempted returns true if the CorrelationID has been attempted before.
// It also marks the CorrelationID as cancelled.
func (r RedisStore) MarkCancelledCheckAttempted(ctx context.Context, s string) bool {
	b, _ := r.client.Eval(ctx, `
redis.call('SET', KEYS[1], "1", "EX", "86400")
if redis.call('EXISTS', KEYS[2]) == 1 then
	return 1
end
return 0
`, []string{r.keyer.Key(":", "cancel", s), r.keyer.Key(":", "attempt", s)}).Bool()
	return b
}

// MarkAttemptedCheckCancelled returns true if the CorrelationID has been cancelled before.
// It also marks the CorrelationID as attempted.
func (r RedisStore) MarkAttemptedCheckCancelled(ctx context.Context, s string) bool {
	b, _ := r.client.Eval(ctx, `
redis.call('SET', KEYS[1], "1", "EX", "86400")
if redis.call('EXISTS', KEYS[2]) == 1 then
	return 1
end
return 0
`, []string{r.keyer.Key(":", "attempt", s), r.keyer.Key(":", "cancel", s)}).Bool()
	return b
}

// Lock grabs the lock for the given key. It returns true if the lock is
// successfully acquired. If the lock is not available, this method will block until
// the lock is released or the context expired. In latter case, false is
// returned.
func (r RedisStore) Lock(ctx context.Context, key string) bool {
	var expiration = time.Minute
	if deadline, ok := ctx.Deadline(); ok {
		expiration = deadline.Sub(time.Now())
	}
	for {
		ok, err := r.client.SetNX(ctx, r.keyer.Key(":", "lock", key), "1", expiration).Result()
		if err == nil && ok {
			return true
		}
		if ctx.Err() != nil {
			return false
		}
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			return false
		}
	}
}

// Unlock unlocks the lock named by key.
func (r RedisStore) Unlock(ctx context.Context, key string) {
	r.client.Del(ctx, r.keyer.Key(":", "lock", key))
}

// Once returns true if this method has been called before with the given key. If
// not, it internally set the key as called and
// returns false.
func (r RedisStore) Once(ctx context.Context, key string) bool {
	_, err := r.client.GetSet(ctx, r.keyer.Key(":", "once", key), "1").Result()
	return err != redis.Nil
}
