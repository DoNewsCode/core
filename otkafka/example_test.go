package otkafka_test

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/segmentio/kafka-go"
)

func Example_reader() {
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
			"reader": map[string]interface{}{
				"default": otkafka.ReaderConfig{
					Brokers: brokers,
					Topic:   "example",
				},
			},
			"writer": map[string]interface{}{
				"default": otkafka.WriterConfig{
					Brokers: brokers,
					Topic:   "example",
				},
			},
		},
	}
	c := core.Default(core.WithConfigStack(confmap.Provider(conf, "."), nil))
	c.Provide(otkafka.Providers())
	c.Invoke(func(writer *kafka.Writer) {
		err := writer.WriteMessages(context.Background(), kafka.Message{Value: []byte(`hello`)})
		if err != nil {
			panic(err)
		}
	})
	c.Invoke(func(reader *kafka.Reader) {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Println(string(msg.Value))
	})
	// Output:
	// hello
}
