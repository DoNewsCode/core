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
	var topics = []string{"trace", "test", "example"}

	for _, topic := range topics {
		conn, err := kafka.Dial("tcp", "localhost:9092")
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
			kafka.TopicConfig{
				Topic:             topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		}

		err = controllerConn.CreateTopics(topicConfigs...)
		if err != nil {
			panic(err.Error())
		}
	}
}
