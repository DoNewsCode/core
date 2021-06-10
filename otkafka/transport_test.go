package otkafka

import (
	"context"
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
	tracer := mocktracer.New()

	factory, cleanup := provideWriterFactory(in{
		Tracer: tracer,
		In:     di.In{},
		Conf: config.MapAdapter{"kafka.writer": map[string]WriterConfig{
			"default": {
				Brokers: envDefaultKafkaAddrs,
				Topic:   "Test",
			},
		}},
		Logger: log.NewNopLogger(),
	})
	defer cleanup()
	def, _ := factory.Make("default")

	span, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "test")
	def.WriteMessages(ctx, kafka.Message{Value: []byte(`foo`)})
	assert.Len(t, tracer.FinishedSpans(), 1)
	span.Finish()

}
