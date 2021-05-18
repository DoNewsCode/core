package processor

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/internal"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/segmentio/kafka-go"
)

var envDefaultKafkaAddrs, envDefaultKafkaAddrsIsSet = internal.GetDefaultAddrsFromEnv("KAFKA_ADDR", "127.0.0.1:9092")

func TestMain(m *testing.M) {
	if !envDefaultKafkaAddrsIsSet {
		fmt.Println("Set env KAFKA_ADDR to run otkafka.processor tests")
		os.Exit(0)
	}
	setupTopic()
	setupMessage()

	os.Exit(m.Run())
}

func setupTopic() {
	conn, err := kafka.Dial("tcp", envDefaultKafkaAddrs[0])
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		panic(err.Error())
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		panic(err.Error())
	}
	defer controllerConn.Close()

	topics := []string{"processor"}
	topicConfigs := make([]kafka.TopicConfig, len(topics))
	for i, topic := range topics {
		topicConfigs[i] = kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
	}

	err = controllerConn.CreateTopics(topicConfigs...)

	if err != nil {
		panic(err.Error())
	}
}

func setupMessage() {
	c := core.New(
		core.WithInline("kafka.writer.default.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.writer.default.topic", "processor"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
	c.Provide(otkafka.Providers())

	c.Invoke(func(w *kafka.Writer) {
		testMessages := make([]kafka.Message, 4)
		for i := 0; i < 4; i++ {
			testMessages[i] = kafka.Message{Value: []byte(fmt.Sprintf(`{"id":%d}`, i))}
		}
		err := w.WriteMessages(context.Background(), testMessages...)
		if err != nil {
			panic(err)
		}
	})

}
