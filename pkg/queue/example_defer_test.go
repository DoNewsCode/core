package queue_test

import (
	"context"
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

type DeferMockData struct {
	Value string
}

type DeferMockListener struct{}

func (m DeferMockListener) Listen() []contract.Event {
	return events.From(DeferMockData{})
}

func (m DeferMockListener) Process(_ context.Context, event contract.Event) error {
	fmt.Println(event.Data().(DeferMockData).Value)
	return nil
}

// bootstrapMetrics is normally done when bootstrapping the framework. We mimic it here for demonstration.
func bootstrapDefer() *core.C {
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
func serveDefer(c *core.C, duration time.Duration) {
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

func Example_defer() {
	c := bootstrapDefer()

	err := c.Invoke(func(dispatcher queue.Dispatcher) {
		// Subscribe
		dispatcher.Subscribe(DeferMockListener{})

		// Trigger an event
		evt := events.Of(DeferMockData{Value: "hello world"})
		_ = dispatcher.Dispatch(context.Background(), queue.Persist(evt, queue.Defer(time.Second)))
		_ = dispatcher.Dispatch(context.Background(), queue.Persist(evt, queue.Defer(time.Hour)))
	})
	if err != nil {
		panic(err)
	}

	serveDefer(c, 2*time.Second)

	// Output:
	// hello world
}
