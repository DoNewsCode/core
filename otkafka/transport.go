package otkafka

import (
	"context"
	"net"

	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
)

// Transport is a type which traces the interacting with kafka brokers.
type Transport struct {
	underlying kafka.RoundTripper
	tracer     opentracing.Tracer
}

// NewTransport creates a new kafka transport
func NewTransport(underlying kafka.RoundTripper, tracer opentracing.Tracer) *Transport {
	return &Transport{
		underlying: underlying,
		tracer:     tracer,
	}
}

// RoundTrip implements kafka.RoundTripper factoryIn kafka-go. It wraps the original
// kafka.RoundTripper and adds a tracing span to it.
func (t *Transport) RoundTrip(ctx context.Context, addr net.Addr, request kafka.Request) (kafka.Response, error) {
	if opentracing.SpanFromContext(ctx) == nil {
		return t.underlying.RoundTrip(ctx, addr, request)
	}
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, t.tracer, "kafka producer")
	defer span.Finish()
	resp, err := t.underlying.RoundTrip(ctx, addr, request)
	if err != nil {
		span.LogKV("error", err.Error())
		span.SetTag("error", err.Error())
	}
	return resp, err
}
