package franz_go

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/events"
	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/kgo"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
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
				Topics:      []string{"Test"},
			},
			"alternative": {
				SeedBrokers: addrs,
				Topics:      []string{"Test"},
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
						Topics:      []string{"Test"},
					},
					"alternative": {
						SeedBrokers: nil,
						Topics:      []string{"Test"},
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

func TestRun(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestProvideFactory")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	factory, cleanup := provideFactory(factoryIn{
		Logger: log.NewJSONLogger(os.Stdout),
		Conf: config.MapAdapter{"kafka": map[string]Config{
			"default": {
				SeedBrokers: addrs,
				Topics:      []string{"test"},
				Group:       "franz-test",
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
	record := &kgo.Record{Topic: "foo", Value: []byte("bar")}
	cli.Produce(ctx, record, func(_ *kgo.Record, err error) {
		defer wg.Done()
		if err != nil {
			t.Fatalf("record had a produce error: %v\n", err)
		}
	})
	wg.Wait()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	for {
		fetches := cli.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			// All errors are retried internally when fetching, but non-retriable errors are
			// returned from polls so that users can notice and take action.
			panic(fmt.Sprint(errs))
		}

		// We can iterate through a record iterator...
		iter := fetches.RecordIter()
		for {
			if iter.Done() {
				return
			}
			record := iter.Next()
			fmt.Println(string(record.Value), "from an iterator!")
		}
	}
}
