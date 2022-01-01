package otkafka

import (
	"context"
	"testing"

	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestHelper_no_parent(t *testing.T) {
	span, _, err := SpanFromMessage(context.Background(), mocktracer.New(), &kafka.Message{})
	assert.NoError(t, err)
	assert.Zero(t, span.(*mocktracer.MockSpan).ParentID)
}
