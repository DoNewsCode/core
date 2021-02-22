package kitkafka

import (
	"context"
	"github.com/segmentio/kafka-go"
)

// RequestResponseFunc may take information from a publisher request and put it into a
// request context. In Subscribers, RequestResponseFunc are executed prior to invoking
// the endpoint.
type RequestResponseFunc func(context.Context, *kafka.Message) context.Context
