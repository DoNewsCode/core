package kafka_go

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestTransport_RoundTrip(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestTransport_RoundTrip")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")

	tracer := mocktracer.New()

	factory, cleanup := provideWriterFactory(factoryIn{
		Tracer: tracer,
		In:     di.In{},
		Conf: config.MapAdapter{"kafka.writer": map[string]WriterConfig{
			"default": {
				Brokers: addrs,
				Topic:   "Test",
			},
		}},
		Logger: log.NewNopLogger(),
	}, func(name string, writer *kafka.Writer) {})
	defer cleanup()
	def, _ := factory.Make("default")

	span, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "test")
	def.WriteMessages(ctx, kafka.Message{Value: []byte(`foo`)})
	assert.Len(t, tracer.FinishedSpans(), 1)
	span.Finish()
}
