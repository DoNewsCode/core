// +build integration

package otkafka

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriter(t *testing.T) {
	{
		ctx := context.Background()
		kw := kafka.Writer{
			Addr:  kafka.TCP("localhost:9092"),
			Topic: "trace",
		}
		tracer := mocktracer.New()
		w := Trace(&kw, tracer, WithLogger(log.NewNopLogger()))
		span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "test")
		span.SetBaggageItem("foo", "bar")
		err := w.WriteMessages(ctx, kafka.Message{Value: []byte(`hello`)})
		assert.NoError(t, err)
		assert.Len(t, tracer.FinishedSpans(), 1)
		span.Finish()
	}

	{
		ctx := context.Background()
		kr := kafka.NewReader(kafka.ReaderConfig{Brokers: []string{"localhost:9092"}, Topic: "trace", GroupID: "test", MinBytes: 1, MaxBytes: 1})
		tracer := mocktracer.New()
		msg, err := kr.ReadMessage(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "hello", string(msg.Value))
		span, ctx, err := SpanFromMessage(ctx, tracer, &msg)
		assert.NoError(t, err)
		foo := span.BaggageItem("foo")
		assert.Equal(t, "bar", foo)
		span.Finish()
	}
}
