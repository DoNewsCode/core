package queue_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/core"
	"github.com/DoNewsCode/std/pkg/events"
	"github.com/DoNewsCode/std/pkg/queue"
	"github.com/go-redis/redis/v8"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/oklog/run"
	"time"
)

type FaultyMockData struct {
	Value string
}

type FaultyMockListener struct {
	count int
}

func (m *FaultyMockListener) Listen() []contract.Event {
	return events.From(FaultyMockData{})
}

func (m *FaultyMockListener) Process(_ context.Context, event contract.Event) error {
	if m.count < 2 {
		fmt.Println("faulty")
		m.count++
		return errors.New("faulty")
	}
	fmt.Println(event.Data().(FaultyMockData).Value)
	return nil
}

// bootstrapMetrics is normally done when bootstrapping the framework. We mimic it here for demonstration.
func bootstrapRetry() *core.C {
	const sampleConfig = "{\"log\":{\"level\":\"error\"},\"queue\":{\"default\":{\"parallelism\":1}}}"

	// Make sure redis is running at localhost:6379
	c := core.New(
		core.WithConfigStack(rawbytes.Provider([]byte(sampleConfig)), json.Parser()),
	)

	// Add Provider
	c.AddCoreDependencies()
	c.AddDependencyFunc(queue.ProvideDispatcher)
	c.AddDependencyFunc(func() redis.UniversalClient {
		client := redis.NewUniversalClient(&redis.UniversalOptions{})
		_, _ = client.FlushAll(context.Background()).Result()
		return client
	})
	return c
}

// serveMetrics normally lives at serveMetrics command. We mimic it here for demonstration.
func serveRetry(c *core.C, duration time.Duration) {
	var g run.Group

	for _, r := range c.GetRunProviders() {
		r(&g)
	}

	// cancel the run group after some time, so that the program ends. In real project, this is not necessary.
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	g.Add(func() error {
		<-ctx.Done()
		return nil
	}, func(err error) {
		cancel()
	})

	err := g.Run()
	if err != nil {
		panic(err)
	}
}

func Example_faulty() {
	c := bootstrapRetry()

	err := c.Invoke(func(dispatcher queue.Dispatcher) {
		// Subscribe
		dispatcher.Subscribe(&FaultyMockListener{})

		// Trigger an event
		evt := events.Of(FaultyMockData{Value: "hello world"})
		_ = dispatcher.Dispatch(context.Background(), queue.Persist(evt, queue.MaxAttempts(3)))
	})
	if err != nil {
		panic(err)
	}

	serveRetry(c, 5*time.Second) // retries are made after a random backoff. It may take longer.

	// Output:
	// faulty
	// faulty
	// hello world
}
