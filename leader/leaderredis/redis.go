// Package leaderredis provides a redis driver for package leader
package leaderredis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/go-redis/redis/v8"
)

// RedisDriver is a simple redis leader election implementation.
type RedisDriver struct {
	client       redis.UniversalClient
	keyer        contract.Keyer
	expiration   time.Duration
	pollInterval time.Duration
	cancel       func()
	sha          string
}

// Option is type of the options to config *RedisDriver
type Option func(driver *RedisDriver)

// WithExpiration is an option that configures the expiration of redis key. A new
// round of leader election will start after the key expires.
func WithExpiration(duration time.Duration) Option {
	return func(driver *RedisDriver) {
		driver.expiration = duration
	}
}

// WithPollInterval is an option that configures the poll interval of followers.
// The followers will periodically try to overthrown the leader, but only succeed
// when the leader key is missing in redis.
func WithPollInterval(duration time.Duration) Option {
	return func(driver *RedisDriver) {
		driver.pollInterval = duration
	}
}

// NewRedisDriver creates the newly created *RedisDriver with the given configuration.
func NewRedisDriver(client redis.UniversalClient, keyer contract.Keyer, opts ...Option) *RedisDriver {
	driver := &RedisDriver{
		client:       client,
		keyer:        keyer,
		expiration:   time.Minute,
		pollInterval: time.Second,
		sha:          "",
	}
	for _, f := range opts {
		f(driver)
	}
	return driver
}

// Campaign starts the leader election using redis. It will bock until this node becomes leader or the context is expired.
func (r *RedisDriver) Campaign(ctx context.Context) error {
	for {
		hostname, _ := os.Hostname()
		ok, err := r.client.SetNX(ctx, r.keyer.Key(":", "leader"), hostname, r.expiration).Result()
		if err != redis.Nil && err != nil {
			return fmt.Errorf("error when running compaign: %w", err)
		}
		if ok {
			var ctx context.Context
			ctx, r.cancel = context.WithCancel(context.Background())
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case <-time.After(1 * r.expiration / 4):
						r.client.Expire(ctx, r.keyer.Key(":", "leader"), r.expiration).Result()
					}
				}
			}()
			return nil
		}
		time.Sleep(r.pollInterval)
	}
}

// Resign gives up the leadership using redis. If the current node is not a leader, this is an no op.
func (r *RedisDriver) Resign(ctx context.Context) error {
	if r.cancel != nil {
		r.cancel()
	}
	hostname, _ := os.Hostname()
	if r.sha == "" {
		var err error
		r.sha, err = r.client.ScriptLoad(context.Background(), `
if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end
`).Result()
		if err != nil {
			return fmt.Errorf("unable to resign: %w", err)
		}
	}
	_, err := r.client.EvalSha(ctx, r.sha, []string{r.keyer.Key(":", "leader")}, hostname).Result()
	if err != nil {
		return fmt.Errorf("unable to resign: %w", err)
	}
	return nil

}
