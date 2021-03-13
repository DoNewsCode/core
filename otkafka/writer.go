package otkafka

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/segmentio/kafka-go"
)

// Writer is a decorator around kafka.Writer that provides tracing capabilities.
type Writer struct {
	*kafka.Writer
	tracer opentracing.Tracer
	logger log.Logger
}

// WriterOption is type that configures the Writer.
type WriterOption func(writer *Writer)

// WithLogger is an option that provides logging to writer.
func WithLogger(logger log.Logger) WriterOption {
	return func(writer *Writer) {
		writer.logger = logger
	}
}

// Trace takes a kafka.Writer and returns a decorated Writer.
func Trace(writer *kafka.Writer, tracer opentracing.Tracer, opts ...WriterOption) *Writer {
	w := &Writer{
		Writer: writer,
		tracer: tracer,
	}
	for _, f := range opts {
		f(w)
	}
	return w
}

// WriteMessages writes a batch of messages to the kafka topic configured on this writer.
// Each message written has tracing headers.
func (w *Writer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, w.tracer, "kafka writer")
	defer span.Finish()

	ext.SpanKind.Set(span, ext.SpanKindProducerEnum)

	carrier := make(opentracing.TextMapCarrier)
	err := w.tracer.Inject(span.Context(), opentracing.TextMap, carrier)
	if err != nil && w.logger != nil {
		_ = level.Warn(w.logger).Log("err", fmt.Sprintf("unable to inject tracing context: %s", err.Error()))
	} else {
		_ = level.Debug(w.logger).Log("msg", fmt.Sprintf("trace injected"))
	}

	for i := range msgs {
		for k := range carrier {
			var header kafka.Header
			header.Key = k
			header.Value = []byte(carrier[k])
			msgs[i].Headers = append(msgs[i].Headers, header)
		}
	}

	err = w.Writer.WriteMessages(ctx, msgs...)
	if err != nil {
		span.SetTag("Error", true)
		span.LogKV("error", "")
	}
	return err
}
