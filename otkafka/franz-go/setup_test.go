package franz_go

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

const franzTestTopic = "franz-test"

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
	adm, err := kgo.NewClient(kgo.SeedBrokers(strings.Split(addr, ",")...))
	if err != nil {
		panic(fmt.Sprintf("unable to create admin client: %v", err))
	}
	topic := franzTestTopic

	req := kmsg.NewPtrCreateTopicsRequest()
	reqTopic := kmsg.NewCreateTopicsRequestTopic()
	reqTopic.Topic = topic
	reqTopic.NumPartitions = 1
	reqTopic.ReplicationFactor = 1
	req.Topics = append(req.Topics, reqTopic)

	resp, err := req.RequestWith(context.Background(), adm)
	if err == nil {
		err = kerr.ErrorForCode(resp.Topics[0].ErrorCode)
	}
	if err != nil {
		if !errors.Is(err, kerr.TopicAlreadyExists) {
			panic(err)
		}
	}

	return func() {
		req := kmsg.NewPtrDeleteTopicsRequest()
		req.TopicNames = []string{topic}
		reqTopic := kmsg.NewDeleteTopicsRequestTopic()
		reqTopic.Topic = kmsg.StringPtr(topic)
		req.Topics = append(req.Topics, reqTopic)

		resp, err := req.RequestWith(context.Background(), adm)
		if err == nil {
			err = kerr.ErrorForCode(resp.Topics[0].ErrorCode)
		}
		if err != nil {
			panic(err)
		}
	}
}
