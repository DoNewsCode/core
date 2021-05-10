package otkafka_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/segmentio/kafka-go"
)

func Example_reader() {
	brokers := make([]string, len(config.ENV_DEFAULT_KAFKA_ADDRS))
	for i, addr := range config.ENV_DEFAULT_KAFKA_ADDRS {
		brokers[i] = fmt.Sprintf(`        - %s`, addr)
	}
	brokersStr := strings.Join(brokers, `
`)
	var conf = `
log:
  level: none
kafka:
  reader:
    default:
      brokers:
` + brokersStr + `
      topic:
        example
  writer:
    default:
      brokers:
` + brokersStr + `
      topic:
        example
`
	c := core.Default(core.WithConfigStack(rawbytes.Provider([]byte(conf)), yaml.Parser()))
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
