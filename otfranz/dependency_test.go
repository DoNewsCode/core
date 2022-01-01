package otfranz

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/events"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}

func TestProvideFactory(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestProvideFactory")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	factory, cleanup := provideFactory(factoryIn{
		Conf: config.MapAdapter{"kafka": map[string]Config{
			"default": {
				SeedBrokers: addrs,
				Topics:      []string{franzTestTopic},
			},
			"alternative": {
				SeedBrokers: addrs,
				Topics:      []string{franzTestTopic},
			},
		}},
	}, func(name string, config *Config) {})
	def, err := factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := factory.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	assert.NotNil(t, cleanup)
	cleanup()
}

func TestProvideKafka(t *testing.T) {
	for _, c := range []struct {
		name       string
		reloadable bool
	}{
		{"reload", true},
		{"not reload", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &events.SyncDispatcher{}
			Out, cleanup, err := provideKafkaFactory(&providersOption{
				reloadable: c.reloadable,
			})(factoryIn{
				Logger: log.NewNopLogger(),
				Conf: config.MapAdapter{"kafka": map[string]Config{
					"default": {
						SeedBrokers: nil,
						Topics:      []string{franzTestTopic},
					},
					"alternative": {
						SeedBrokers: nil,
						Topics:      []string{franzTestTopic},
					},
				}},
				Dispatcher: dispatcher,
			})
			assert.NoError(t, err)
			def, err := Out.Factory.Make("default")
			assert.NoError(t, err)
			assert.NotNil(t, def)
			alt, err := Out.Factory.Make("alternative")
			assert.NoError(t, err)
			assert.NotNil(t, alt)
			assert.Equal(t, c.reloadable, dispatcher.ListenerCount(events.OnReload) == 1)
			cleanup()
		})
	}
}

func TestProduceAndConsume(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestProvideFactory")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	factory, cleanup := provideFactory(factoryIn{
		Logger: log.NewNopLogger(),
		Conf: config.MapAdapter{"kafka": map[string]Config{
			"default": {
				SeedBrokers:         addrs,
				DefaultProduceTopic: franzTestTopic,
				Topics:              []string{franzTestTopic},
				Group:               "franz-test",
			},
		}},
	}, func(name string, config *Config) {})
	defer cleanup()
	cli, err := factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(1)
	record := &kgo.Record{Value: []byte("bar")}
	cli.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			t.Fatalf("record had a produce error: %v\n", err)
		}
	})
	wg.Wait()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fetches := cli.PollFetches(ctx)
	if errs := fetches.Errors(); len(errs) > 0 {
		// All errors are retried internally when fetching, but non-retriable errors are
		// returned from polls so that users can notice and take action.
		t.Fatal(errs)
	}

	// We can iterate through a record iterator...
	iter := fetches.RecordIter()
	if iter.Done() {
		t.Fatal("no message consumed")
	}
	rec := iter.Next()
	assert.Equal(t, record.Value, rec.Value)
}
