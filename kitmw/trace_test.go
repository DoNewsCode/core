package kitmw

import (
	"context"
	"errors"
	"testing"

	"github.com/DoNewsCode/core/key"
	"github.com/go-kit/kit/endpoint"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

func TestTraceServer(t *testing.T) {
	var (
		original endpoint.Endpoint
		wrapped  endpoint.Endpoint
		tracer   *mocktracer.MockTracer
	)
	tracer = mocktracer.New()
	original = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return nil, errors.New("")
	}
	wrapped = TraceServer(tracer, key.New())(original)
	wrapped(context.Background(), nil)
	assert.NotEmpty(t, tracer.FinishedSpans())
}
