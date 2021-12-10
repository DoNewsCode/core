package leaderredis

import (
	"context"
	"os"
	"strings"
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
	driver.Campaign(context.Background())

	_, err := client.Get(context.Background(), "leader").Result()
	assert.NoError(t, err)

	driver.Resign(context.Background())

	_, err = client.Get(context.Background(), "leader").Result()
	assert.Error(t, err)
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
	dispatcher := &events.SyncDispatcher{}
	var e1, e2 *leader.Election
	driver := NewRedisDriver(redis.NewUniversalClient(&redis.UniversalOptions{Addrs: addrs}), key.New("testElection"), WithPollInterval(time.Millisecond), WithExpiration(time.Second))

	e1 = leader.NewElection(dispatcher, driver)
	e2 = leader.NewElection(dispatcher, driver)
	ctx, cancel := context.WithCancel(context.Background())

	e1.Campaign(ctx)
	assert.Equal(t, e1.Status().IsLeader(), true)
	go e2.Campaign(ctx)
	<-time.After(2 * time.Second)

	assert.Equal(t, true, e1.Status().IsLeader())
	assert.Equal(t, false, e2.Status().IsLeader())

	e1.Resign(ctx)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, e1.Status().IsLeader())
	assert.Equal(t, true, e2.Status().IsLeader())
	e2.Resign(ctx)
	cancel()
}
