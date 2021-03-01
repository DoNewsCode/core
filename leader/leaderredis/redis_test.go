// +build integration

package leaderredis

import (
	"context"
	"testing"

	"github.com/DoNewsCode/core/key"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestCampaign(t *testing.T) {
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
