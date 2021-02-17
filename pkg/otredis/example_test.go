package otredis_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/std/pkg/core"
	"github.com/DoNewsCode/std/pkg/otredis"
	"github.com/go-redis/redis/v8"
)

func Example() {
	c := core.New()
	c.AddCoreDependencies()
	c.AddDependencyFunc(otredis.ProvideRedis)
	c.Invoke(func(redisClient redis.UniversalClient) {
		pong, _ := redisClient.Ping(context.Background()).Result()
		fmt.Println(pong)
	})
	// Output:
	// PONG
}
