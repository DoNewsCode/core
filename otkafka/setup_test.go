package otkafka

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"net"
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("KAFKA_ADDR") == "" {
		fmt.Println("Set env KAFKA_ADDR to run otkafka tests")
		os.Exit(0)
	}

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
