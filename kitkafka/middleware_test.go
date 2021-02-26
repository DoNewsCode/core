package kitkafka

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestContextToKafka(t *testing.T) {
	tracer := mocktracer.New()
	logger := log.NewNopLogger()
	f := ContextToKafka(tracer, logger)

	ctx := context.Background()
	message := &kafka.Message{}
	f(ctx, message)
	assert.Empty(t, message.Headers)

	_, ctx = opentracing.StartSpanFromContextWithTracer(ctx, tracer, "test")
	message = &kafka.Message{}
	f(ctx, message)
	assert.NotEmpty(t, message.Headers)
}

func TestKafkaToContext_nospan(t *testing.T) {
	tracer := mocktracer.New()
	logger := log.NewNopLogger()
	f := KafkaToContext(tracer, "test", logger)

	ctx := context.Background()
	message := &kafka.Message{}
	ctx = f(ctx, message)
	span := opentracing.SpanFromContext(ctx)
	assert.Equal(t, "test", span.(*mocktracer.MockSpan).OperationName)
	assert.Equal(t, 0, span.(*mocktracer.MockSpan).ParentID)
}

func TestKafkaToContext_withspan(t *testing.T) {
	tracer := mocktracer.New()
	tracer.RegisterExtractor(opentracing.TextMap, &mocktracer.TextMapPropagator{})
	tracer.RegisterInjector(opentracing.TextMap, &mocktracer.TextMapPropagator{})
	opentracing.SetGlobalTracer(tracer)
	logger := log.NewNopLogger()
	f := KafkaToContext(tracer, "test", logger)

	ctx := context.Background()
	ctk := ContextToKafka(tracer, logger)
	message := &kafka.Message{}
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "test")
	span.SetBaggageItem("foo", "bar")
	spanId := span.(*mocktracer.MockSpan).SpanContext.SpanID
	traceId := span.(*mocktracer.MockSpan).SpanContext.TraceID
	ctk(ctx, message)

	ctx = f(context.Background(), message)
	span = opentracing.SpanFromContext(ctx)
	//assert.Equal(t, "test", span.(*mocktracer.MockSpan).OperationName)
	assert.Equal(t, traceId, span.(*mocktracer.MockSpan).SpanContext.TraceID)
	assert.Equal(t, spanId, span.(*mocktracer.MockSpan).ParentID)
	assert.Equal(t, "bar", span.BaggageItem("foo"))
}

func TestErrHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := log.NewLogfmtLogger(&buf)
	errHandler := ErrHandler(logger)
	errHandler.Handle(context.Background(), errors.New("foo"))
	assert.Equal(t, "level=warn err=foo\n", buf.String())
}
