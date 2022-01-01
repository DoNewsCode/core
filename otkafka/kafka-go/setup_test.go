package kafka_go

import (
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestMain(m *testing.M) {
	var cleanup func()
	if os.Getenv("KAFKA_ADDR") != "" {
		cleanup = setupTopic(os.Getenv("KAFKA_ADDR"))
	}
	code := m.Run()
	if cleanup != nil {
		cleanup()
	}
	os.Exit(code)
}

func setupTopic(addr string) func() {
	topics := []string{"trace", "test", "example"}

	conn, err := kafka.Dial("tcp", addr)
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

	topicConfigs := make([]kafka.TopicConfig, 0)
	for _, topic := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}

	return func() {
		defer controllerConn.Close()
		err := controllerConn.DeleteTopics(topics...)
		if err != nil {
			panic(err)
		}
	}
}
