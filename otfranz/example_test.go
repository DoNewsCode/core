package otfranz_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otfranz"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/twmb/franz-go/pkg/kgo"
)

func Example() {
	if os.Getenv("KAFKA_ADDR") == "" {
		fmt.Println("set KAFKA_ADDR to run this example")
		return
	}
	brokers := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	conf := map[string]interface{}{
		"log": map[string]interface{}{
			"level": "none",
		},
		"kafka": map[string]interface{}{
			"default": otfranz.Config{
				SeedBrokers:            brokers,
				DefaultProduceTopic:    "franz-example",
				AllowAutoTopicCreation: true,
				Topics:                 []string{"franz-example"},
				Group:                  "franz-test",
			},
		},
	}
	c := core.Default(core.WithConfigStack(confmap.Provider(conf, "."), nil))
	c.Provide(otfranz.Providers())

	c.Invoke(func(cli *kgo.Client) {
		record := &kgo.Record{Value: []byte("bar")}
		cli.Produce(context.Background(), record, nil)
	})

	c.Invoke(func(cli *kgo.Client) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fetches := cli.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			panic(errs)
		}
		iter := fetches.RecordIter()
		if iter.Done() {
			return
		}
		rec := iter.Next()
		fmt.Println(string(rec.Value))
	})

	// Output:
	// bar
}
