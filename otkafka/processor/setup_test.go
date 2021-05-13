package processor

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/DoNewsCode/core/internal"
	"github.com/segmentio/kafka-go"
)

var envDefaultKafkaAddrs, envDefaultKafkaAddrsIsSet = internal.GetDefaultAddrsFromEnv("KAFKA_ADDR", "127.0.0.1:9092")

func TestMain(m *testing.M) {
	if !envDefaultKafkaAddrsIsSet {
		fmt.Println("Set env KAFKA_ADDR to run otkafka.processor tests")
		os.Exit(0)
	}
	setupTopic()

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

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             "processor1",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             "processor2",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)

	if err != nil {
		panic(err.Error())
	}
}
