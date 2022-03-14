package leaderredis

import (
	"context"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/key"
	"github.com/DoNewsCode/core/leader"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestCampaign(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("set REDIS_ADDR to run TestCampaign")
		return
	}
	addrs := strings.Split(os.Getenv("REDIS_ADDR"), ",")
	client := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: addrs})
	driver := RedisDriver{
		client: client,
		keyer:  key.New(),
	}
	status := &atomic.Value{}
	status.Store(false)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go driver.Campaign(ctx, func(b bool) {
		status.Store(b)
	})

	assert.Eventually(t, func() bool {
		return status.Load().(bool) == true
	}, time.Second, 10*time.Millisecond, "campaign should have started and successfully become leader")

	driver.Resign(ctx)

	assert.Eventually(t, func() bool {
		return status.Load().(bool) == false
	}, time.Second, 10*time.Millisecond, "leader should have resigned")
}

func TestNewRedisDriver(t *testing.T) {
	client := redis.NewUniversalClient(&redis.UniversalOptions{})
	driver := NewRedisDriver(client, key.New("foo", "bar"), WithExpiration(time.Hour), WithPollInterval(time.Minute))
	assert.NotNil(t, driver)
}

func TestElection(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("set REDIS_ADDR to run TestCampaign")
		return
	}
	addrs := strings.Split(os.Getenv("REDIS_ADDR"), ",")
	dispatcher := &events.Event[*leader.Status]{}
	var e1, e2 *leader.Election
	driver := NewRedisDriver(redis.NewUniversalClient(&redis.UniversalOptions{Addrs: addrs}), key.New("testElection"), WithPollInterval(time.Millisecond), WithExpiration(time.Second))

	e1 = leader.NewElection(dispatcher, driver)
	e2 = leader.NewElection(dispatcher, driver)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go e1.Campaign(ctx)
	assert.Eventually(t, func() bool {
		return e1.Status().IsLeader()
	}, time.Second, 10*time.Millisecond, "e1 should be leader")

	go e2.Campaign(ctx)

	assert.Never(t, func() bool {
		return e2.Status().IsLeader()
	}, time.Second, 10*time.Millisecond, "e2 should not be leader")

	e1.Resign(ctx)

	assert.Eventually(t, func() bool {
		return e2.Status().IsLeader()
	}, time.Second, 10*time.Millisecond, "e2 should be leader")
	assert.Never(t, func() bool {
		return e1.Status().IsLeader()
	}, time.Second, 10*time.Millisecond, "e1 should not be leader")

	e2.Resign(ctx)
	cancel()
}
