package kitkafka

import (
	"context"
	"flag"
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/stretchr/testify/assert"
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

	kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "Test", 0)

	writerFactory, cleanupWriter := ProvideKafkaWriterFactory(KafkaIn{
		Conf: config.MapAdapter{"kafka.writer": map[string]kafka.Writer{
			"default": {
				Addr:  kafka.TCP("127.0.0.1:9092"),
				Topic: "Test",
			},
		}},
		Logger: logging.NewLogger("logfmt"),
	})
	defer cleanupWriter()

	readerFactory, cleanupReader := ProvideKafkaReaderFactory(KafkaIn{
		Conf: config.MapAdapter{"kafka.reader": map[string]kafka.ReaderConfig{
			"default": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "Test",
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
