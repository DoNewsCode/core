package leaderredis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// RedisDriver is a simple redis leader election implementation.
type RedisDriver struct {
	client redis.UniversalClient
	keyer  contract.Keyer
}

// Campaign starts the leader election using redis. It will bock until this node becomes leader or the context is expired.
func (r *RedisDriver) Campaign(ctx context.Context) error {
	for {
		hostname, _ := os.Hostname()
		ok, err := r.client.SetNX(ctx, r.keyer.Key(":", "leader"), hostname, time.Minute).Result()
		if err != redis.Nil && err != nil {
			return fmt.Errorf("error when running compaign: %w", err)
		}
		if ok {
			return nil
		}
		time.Sleep(time.Second)
	}
}

// Resign gives up the leadership using redis. If the current node is not a leader, this is an no op.
func (r RedisDriver) Resign(ctx context.Context) error {
	hostname, _ := os.Hostname()
	// TODO: make read and delete atomic
	leader, _ := r.client.Get(ctx, r.keyer.Key(":", "leader")).Result()
	if hostname == leader {
		r.client.Del(ctx, r.keyer.Key(":", "leader")).Result()
		return nil
	}
	return errors.New("not leader")

}
