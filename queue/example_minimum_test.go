package queue_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/queue"
	"sync"
	"time"
)

func Example_minimum() {
	dispatcher := events.SyncDispatcher{}
	queueDispatcher := queue.WithQueue(&dispatcher, queue.NewInProcessDriver())
	ctx, cancel := context.WithCancel(context.Background())
	queueDispatcher.Dispatch(ctx, queue.Persist(events.Of(1), queue.Defer(time.Second)))
	queueDispatcher.Dispatch(ctx, queue.Persist(events.Of(2), queue.Defer(time.Hour)))
	var wg sync.WaitGroup
	wg.Add(1)
	go queueDispatcher.Consume(ctx)
	queueDispatcher.Subscribe(events.Listen(events.From(1), func(ctx context.Context, event contract.Event) error {
		fmt.Println(event.Data())
		wg.Done()
		return nil
	}))
	wg.Wait()
	cancel()

	// Output:
	// 1
}
