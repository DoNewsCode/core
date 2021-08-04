package otkafka

import (
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestMain(m *testing.M) {
	if os.Getenv("KAFKA_ADDR") != "" {
		setupTopic(os.Getenv("KAFKA_ADDR"))
	}
	os.Exit(m.Run())
}

func setupTopic(addr string) {
	var topics = []string{"trace", "test", "example"}

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
	defer controllerConn.Close()

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
}
