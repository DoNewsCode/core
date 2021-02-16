package queue_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/core"
	"github.com/DoNewsCode/std/pkg/events"
	"github.com/DoNewsCode/std/pkg/queue"
	"github.com/DoNewsCode/std/pkg/queue/modqueue"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-redis/redis/v8"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/oklog/run"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"time"
)

type MockData struct {
	Value string
}

type MockListener struct{}

func (m MockListener) Listen() []contract.Event {
	return events.From(MockData{})
}

func (m MockListener) Process(_ context.Context, event contract.Event) error {
	fmt.Println(event.Data().(MockData).Value)
	return nil
}

// bootstrap is normally done when bootstrapping the framework. We mimic it here for demonstration.
func bootstrap() *core.C {
	const sampleConfig = "{\"log\":{\"level\":\"error\"},\"queue\":{\"default\":{\"parallelism\":1}}}"

	// Make sure redis is running at localhost:6379
	c := core.New(
		core.WithConfigStack(rawbytes.Provider([]byte(sampleConfig)), json.Parser()),
	)

	// Add Provider
	c.AddCoreDependencies()
	c.AddDependency(modqueue.ProvideDispatcher)
	c.AddDependency(func() redis.UniversalClient {
		client := redis.NewUniversalClient(&redis.UniversalOptions{})
		_, _ = client.FlushAll(context.Background()).Result()
		return client
	})
	c.AddDependency(func(appName contract.AppName, env contract.Env) modqueue.Gauge {
		return prometheus.NewGaugeFrom(
			stdprometheus.GaugeOpts{
				Namespace: appName.String(),
				Subsystem: env.String(),
				Name:      "queue_length",
				Help:      "The gauge of queue length",
			}, []string{"name", "channel"},
		)
	})
	return c
}

// serve normally lives at serve command. We mimic it here for demonstration.
func serve(c *core.C, duration time.Duration) {
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

func Example_metrics() {
	c := bootstrap()

	err := c.Invoke(func(dispatcher modqueue.Dispatcher) {
		// Subscribe
		dispatcher.Subscribe(MockListener{})

		// Trigger an event
		evt := events.Of(MockData{Value: "hello world"})
		_ = dispatcher.Dispatch(context.Background(), queue.Persist(evt))
	})
	if err != nil {
		panic(err)
	}

	serve(c, time.Second)

	// Output:
	// hello world
}
