// +build integration

package leaderredis

import (
	"context"
	"testing"
	"time"

	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/key"
	"github.com/DoNewsCode/core/leader"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestCampaign(t *testing.T) {
	t.Parallel()
	client := redis.NewUniversalClient(&redis.UniversalOptions{})
	client.FlushAll(context.Background()).Result()

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
	t.Parallel()
	client := redis.NewUniversalClient(&redis.UniversalOptions{})
	client.FlushAll(context.Background()).Result()

	driver := NewRedisDriver(client, key.New("foo", "bar"), WithExpiration(time.Hour), WithPollInterval(time.Minute))
	assert.NotNil(t, driver)
}

func TestElection(t *testing.T) {
	t.Parallel()
	var dispatcher = &events.SyncDispatcher{}
	var e1, e2 *leader.Election
	var driver = NewRedisDriver(redis.NewUniversalClient(&redis.UniversalOptions{}), key.New("testElection"), WithPollInterval(time.Microsecond), WithExpiration(time.Second))

	e1 = leader.NewElection(dispatcher, driver)
	e2 = leader.NewElection(dispatcher, driver)
	ctx, cancel := context.WithCancel(context.Background())

	e1.Campaign(ctx)
	assert.Equal(t, e1.Status().IsLeader(), true)

	go e2.Campaign(ctx)
	<-time.After(3 * time.Second)

	assert.Equal(t, true, e1.Status().IsLeader())
	assert.Equal(t, false, e2.Status().IsLeader())

	e1.Resign(ctx)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, e1.Status().IsLeader())
	assert.Equal(t, true, e2.Status().IsLeader())
	e2.Resign(ctx)
	cancel()
}
