package kitkafka

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/segmentio/kafka-go"
	jaeger "github.com/uber/jaeger-client-go/config"
)

var useKafka bool

func init() {
	flag.BoolVar(&useKafka, "kafka", false, "use kafka for testing")
}

func TestTransport(t *testing.T) {
	if !useKafka {
		t.Skip("requires kafka")
	}

	factory := NewKafkaFactory([]string{"127.0.0.1:9092"}, log.NewNopLogger())
	h := factory.MakeHandler("test")
	_ = h.Handle(context.Background(), kafka.Message{
		Value: []byte("hello"),
	})
	factory.MakeKafkaServer("test", HandleFunc(func(ctx context.Context, message kafka.Message) error {
		if string(message.Value) != "hello" {
			t.Fatalf("want hello, got %s", message.Value)
		}
		fmt.Println(string(message.Value))
		return nil
	})).ServeOnce(context.Background())
}

func TestTransportTracing(t *testing.T) {
	if !useKafka {
		t.Skip("requires kafka")
	}

	tracer, closer, _ := jaeger.Configuration{
		ServiceName: "your-service-name",
	}.NewTracer()
	defer closer.Close()

	factory := NewKafkaFactory([]string{"127.0.0.1:9092"}, log.NewNopLogger())
	h := factory.MakeHandler("test-tracing")
	h = MakeTracingProducerMiddleware(tracer, "test")(h)

	_ = h.Handle(context.Background(), kafka.Message{
		Value: []byte("hello"),
	})

	sub :=
		HandleFunc(func(ctx context.Context, message kafka.Message) error {
			if message.Headers[0].Key != "uber-trace-id" {
				t.Fatal("context not propagated")
			}
			return nil
		})
	h = MakeTracingConsumerMiddleware(tracer, "test")(sub)

	factory.MakeKafkaServer("test-tracing", h, WithGroup("foo")).ServeOnce(context.Background())
}
