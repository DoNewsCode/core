package otfranz

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Client is a decorator around *kgo.Client that provides tracing capabilities.
type Client struct {
	*kgo.Client
	tracer opentracing.Tracer
}

// NewClient takes a *kgo.Client and returns a decorated Client.
func NewClient(client *kgo.Client, tracer opentracing.Tracer) *Client {
	return &Client{Client: client, tracer: tracer}
}

// ProduceWithTracer wrap Produce method with tracing.
func (c *Client) ProduceWithTracer(ctx context.Context, r *kgo.Record, promise func(*kgo.Record, error)) {
	if c.tracer == nil {
		c.Produce(ctx, r, promise)
		return
	}
	if opentracing.SpanFromContext(ctx) == nil {
		c.Produce(ctx, r, promise)
		return
	}
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, c.tracer, "kafka producer")
	defer span.Finish()

	ext.SpanKind.Set(span, ext.SpanKindProducerEnum)

	c.Produce(ctx, r, func(record *kgo.Record, err error) {
		if err != nil {
			ext.LogError(span, err)
		}
		if promise != nil {
			promise(record, err)
		}
	})
}
