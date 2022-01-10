package otfranz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
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

func testTopics(topics ...string) (*kmsg.CreateTopicsRequest, *kmsg.DeleteTopicsRequest) {
	creq := kmsg.NewPtrCreateTopicsRequest()
	dreq := kmsg.NewPtrDeleteTopicsRequest()
	dreq.TopicNames = topics

	for _, topic := range topics {
		creqTopic := kmsg.NewCreateTopicsRequestTopic()
		creqTopic.Topic = topic
		creqTopic.NumPartitions = 1
		creqTopic.ReplicationFactor = 1
		creq.Topics = append(creq.Topics, creqTopic)

		dreqTopic := kmsg.NewDeleteTopicsRequestTopic()
		dreqTopic.Topic = kmsg.StringPtr(topic)
		dreq.Topics = append(dreq.Topics, dreqTopic)
	}
	return creq, dreq
}

func setupTopic(addr string) func() {
	topics := []string{"franz-test", "franz-tracing", "franz-example", "franz-foo", "franz-bar"}

	adm, err := kgo.NewClient(kgo.SeedBrokers(strings.Split(addr, ",")...))
	if err != nil {
		panic(fmt.Sprintf("unable to create admin client: %v", err))
	}

	creq, dreq := testTopics(topics...)

	resp, err := creq.RequestWith(context.Background(), adm)
	if err == nil {
		err = kerr.ErrorForCode(resp.Topics[0].ErrorCode)
	}
	if err != nil {
		if !errors.Is(err, kerr.TopicAlreadyExists) {
			panic(err)
		}
	}

	return func() {
		_, err := dreq.RequestWith(context.Background(), adm)
		if err != nil {
			panic(err)
		}
	}
}
