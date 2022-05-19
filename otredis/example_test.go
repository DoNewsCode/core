package otredis_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otredis"

	"github.com/go-redis/redis/v8"
)

func Example() {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(otredis.Providers())
	c.Invoke(func(redisClient redis.UniversalClient) {
		pong, _ := redisClient.Ping(context.Background()).Result()
		fmt.Println(pong)
	})
	// Output:
	// PONG
}
