package otkafka

import (
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/segmentio/kafka-go"
)

func TestMain(m *testing.M) {
	setupTopic()

	os.Exit(m.Run())
}

func setupTopic() {
	var topics = []string{"trace", "test", "example"}

	conn, err := kafka.Dial("tcp", config.ENV_DEFAULT_KAFKA_ADDRS[0])
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
