package processor

import (
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestMain(m *testing.M) {
	setupTopic()

	os.Exit(m.Run())
}

func setupTopic() {
	conn, err := kafka.Dial("tcp", os.Getenv("KAFKA_ADDR"))
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
			Topic:             "processor",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)

	if err != nil {
		panic(err.Error())
	}
}
