package kafka_go_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core/otkafka/kafka-go"
	"os"
	"strings"

	"github.com/DoNewsCode/core"
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
				"default": kafka_go.ReaderConfig{
					Brokers: brokers,
					Topic:   "example",
				},
			},
			"writer": map[string]interface{}{
				"default": kafka_go.WriterConfig{
					Brokers: brokers,
					Topic:   "example",
				},
			},
		},
	}
	c := core.Default(core.WithConfigStack(confmap.Provider(conf, "."), nil))
	c.Provide(kafka_go.Providers())
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
