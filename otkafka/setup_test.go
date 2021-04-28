package otkafka

import (
	"github.com/segmentio/kafka-go"
	"net"
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	setupTopic()
	os.Exit(m.Run())
}

func setupTopic() {
	conn, err := kafka.Dial("tcp", "127.0.0.1:9092")
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

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             "trace",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "test",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "example",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}
}
