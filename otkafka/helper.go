package otkafka

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/segmentio/kafka-go"
)

// SpanFromMessage reads the message
func SpanFromMessage(ctx context.Context, tracer opentracing.Tracer, message *kafka.Message) (opentracing.Span, context.Context, error) {
	carrier := getCarrier(message)
	spanContext, err := tracer.Extract(opentracing.TextMap, carrier)
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return nil, nil, err
	}
	span := tracer.StartSpan("kafka reader", ext.RPCServerOption(spanContext))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return nil, nil, err
	}
	ext.SpanKind.Set(span, ext.SpanKindConsumerEnum)
	ext.PeerService.Set(span, "kafka")
	span.SetTag("topic", message.Topic)
	span.SetTag("partition", message.Partition)
	span.SetTag("offset", message.Offset)

	return span, opentracing.ContextWithSpan(ctx, span), nil
}

func getCarrier(msg *kafka.Message) opentracing.TextMapCarrier {

	var mapCarrier = make(opentracing.TextMapCarrier)
	if msg.Headers != nil {
		for _, v := range msg.Headers {
			mapCarrier[v.Key] = string(v.Value)
		}
	}
	return mapCarrier
}
