package dtransaction

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	keyer  contract.Keyer
	client redis.UniversalClient
}

func (r RedisStore) MarkCancelledCheckAttempted(ctx context.Context, s string) bool {
	b, _ := r.client.Eval(ctx, `
redis.call('SET', key[1], "1", "EX", "86400")
if redis.call('EXISTS', key[2]) == 1 then
	return 0
end
return 1
`, []string{r.keyer.Key(":", "cancel", s), r.keyer.Key(":", "attempt", s)}).Bool()
	return b
}

func (r RedisStore) MarkAttemptedCheckCancelled(ctx context.Context, s string) bool {
	b, _ := r.client.Eval(ctx, `
redis.call('SET', key[1], "1", "EX", "86400")
if redis.call('EXISTS', key[2]) == 1 then
	return 0
end
return 1
`, []string{r.keyer.Key(":", "attempt", s), r.keyer.Key(":", "cancel", s)}).Bool()
	return b
}

func (r RedisStore) Lock(ctx context.Context, key string) bool {
	var expiration time.Duration = time.Minute
	if deadline, ok := ctx.Deadline(); ok {
		expiration = deadline.Sub(time.Now())
	}
	for {
		err := r.client.SetNX(ctx, r.keyer.Key(":", "lock", key), "1", expiration)
		if ctx.Err() != nil {
			return false
		}
		if err == nil {
			return true
		}
		time.Sleep(time.Second)
	}
}

func (r RedisStore) Unlock(ctx context.Context, key string) {
	r.client.Del(ctx, r.keyer.Key(":", "lock", key))
}

func (r RedisStore) Once(ctx context.Context, key string) bool {
	_, err := r.client.GetSet(ctx, r.keyer.Key(":", "once", key), "1").Result()
	return err != redis.Nil
}
