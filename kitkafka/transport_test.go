package kitkafka

import (
	"context"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/stretchr/testify/assert"

	"github.com/DoNewsCode/core/logging"
	"github.com/segmentio/kafka-go"
)

func TestTransport(t *testing.T) {

	kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "Test", 0)

	writerFactory, cleanupWriter := ProvideWriterFactory(KafkaIn{
		Conf: config.MapAdapter{"kafka.writer": map[string]WriterConfig{
			"default": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "test",
			},
		}},
		Logger: logging.NewLogger("logfmt"),
	})
	defer cleanupWriter()

	readerFactory, cleanupReader := ProvideReaderFactory(KafkaIn{
		Conf: config.MapAdapter{"kafka.reader": map[string]ReaderConfig{
			"default": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "test",
			},
		}},
		Logger: logging.NewLogger("logfmt"),
	})
	defer cleanupReader()

	// write test data
	h, err := writerFactory.MakeClient("default")
	assert.NoError(t, err)

	err = h.Handle(context.Background(), kafka.Message{
		Value: []byte("hello"),
	})
	assert.NoError(t, err)

	// consume test data
	var consumed = false

	endpoint := func(ctx context.Context, message interface{}) (interface{}, error) {
		if message.(string) != "hello" {
			t.Fatalf("want hello, got %s", message)
		}
		consumed = true
		return nil, nil
	}
	sub, err := readerFactory.MakeSubscriberServer("default", NewSubscriber(endpoint, func(ctx context.Context, message *kafka.Message) (request interface{}, err error) {
		return string(message.Value), nil
	}))

	assert.NoError(t, err)

	err = sub.serveOnce(context.Background())

	assert.NoError(t, err)

	if !consumed {
		t.Fatal("failed to consume the message")
	}
}
