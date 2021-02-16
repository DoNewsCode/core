package queue_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/events"
	"github.com/DoNewsCode/std/pkg/queue"
	"github.com/go-redis/redis/v8"
	"time"
)

func Example_minimum() {
	dispatcher := events.SyncDispatcher{}
	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{})
	redisClient.FlushAll(context.Background())
	queueDispatcher := queue.WithQueue(&dispatcher, &queue.RedisDriver{RedisClient: redisClient})
	ctx, cancel := context.WithCancel(context.Background())
	go queueDispatcher.Consume(ctx)
	queueDispatcher.Subscribe(events.Listen(events.From(1), func(ctx context.Context, event contract.Event) error {
		fmt.Println(event.Data())
		return nil
	}))
	queueDispatcher.Dispatch(ctx, queue.Persist(events.Of(1)))
	time.Sleep(time.Second)
	cancel()

	// Output:
	// 1
}
