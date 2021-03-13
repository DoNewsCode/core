package otkafka_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/segmentio/kafka-go"
)

func Example_reader() {
	var config = `
log:
  level: none
kafka:
  reader:
    default:
      brokers:
        - localhost:9092
      topic:
        example
  writer:
    default:
      brokers:
        - localhost:9092
      topic:
        example
`
	c := core.Default(core.WithConfigStack(rawbytes.Provider([]byte(config)), yaml.Parser()))
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
