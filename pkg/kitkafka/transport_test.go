package kitkafka

import (
	"context"
	"flag"
	"testing"

	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/segmentio/kafka-go"
)

var useKafka bool

func init() {
	flag.BoolVar(&useKafka, "kafka", false, "use kafka for testing")
}

func TestTransport(t *testing.T) {
	if !useKafka {
		t.Skip("requires kafka")
	}

	factory := NewKafkaFactory([]string{"localhost:9092"}, logging.NewLogger("logfmt"))

	// write test data
	h := factory.MakeWriterHandle("test")
	err := h.Handle(context.Background(), kafka.Message{
		Value: []byte("hello"),
	})
	if err != nil {
		t.Fatalf("received expected err %s", err)
	}

	// consume test data
	var consumed = false

	endpoint := func(ctx context.Context, message interface{}) (interface{}, error) {
		if message.(string) != "hello" {
			t.Fatalf("want hello, got %s", message)
		}
		consumed = true
		return nil, nil
	}
	err = factory.MakeSubscriberClient("test", NewSubscriber(endpoint, func(ctx context.Context, message *kafka.Message) (request interface{}, err error) {
		return string(message.Value), nil
	})).ServeOnce(context.Background())

	if err != nil {
		t.Fatalf("received expected err %s", err)
	}

	if !consumed {
		t.Fatal("failed to consume the message")
	}
}
