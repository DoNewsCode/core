package queue_test

import (
	"context"
	"sync"
	"testing"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/queue"
)

func setUpInProcessQueueBenchmark(wg *sync.WaitGroup) (*queue.QueueableDispatcher, func()) {
	dispatcher := events.SyncDispatcher{}
	queueDispatcher := queue.WithQueue(&dispatcher, queue.NewInProcessDriver())
	ctx, cancel := context.WithCancel(context.Background())
	go queueDispatcher.Consume(ctx)
	queueDispatcher.Subscribe(events.Listen(events.From(1), func(ctx context.Context, event contract.Event) error {
		wg.Done()
		return nil
	}))
	return queueDispatcher, cancel
}

func setUpRedisQueueBenchmark(wg *sync.WaitGroup) (*queue.QueueableDispatcher, func()) {
	dispatcher := events.SyncDispatcher{}
	queueDispatcher := queue.WithQueue(&dispatcher, &queue.RedisDriver{})
	ctx, cancel := context.WithCancel(context.Background())
	go queueDispatcher.Consume(ctx)
	queueDispatcher.Subscribe(events.Listen(events.From(1), func(ctx context.Context, event contract.Event) error {
		wg.Done()
		return nil
	}))
	return queueDispatcher, cancel
}

func BenchmarkRedisQueue(b *testing.B) {
	var wg sync.WaitGroup
	dispatcher, cancel := setUpRedisQueueBenchmark(&wg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		dispatcher.Dispatch(context.Background(), queue.Persist(events.Of(1)))
	}
	wg.Wait()
	cancel()
}

func BenchmarkInProcessQueue(b *testing.B) {
	var wg sync.WaitGroup
	dispatcher, cancel := setUpInProcessQueueBenchmark(&wg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		dispatcher.Dispatch(context.Background(), queue.Persist(events.Of(1)))
	}
	wg.Wait()
	cancel()
}
