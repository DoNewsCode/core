package kitmw

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

func TestMakeAsyncMiddleware(t *testing.T) {
	var c int
	m := MakeAsyncMiddleware(log.NewNopLogger(), 5)
	f := m(func(ctx context.Context, request interface{}) (response interface{}, err error) {
		c++
		assert.Less(t, c, 5)
		time.Sleep(time.Duration(rand.Float64()) * time.Second)
		c--
		return nil, nil
	})

	for i := 0; i < 10; i++ {
		t.Run("", func(t *testing.T) {
			t.Parallel()
			f(context.Background(), nil)
		})
	}
}

func TestMakeAsyncMiddleware_tracing(t *testing.T) {
	tracer := mocktracer.New()
	span, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "foo")
	var done = make(chan struct{})

	m := MakeAsyncMiddleware(log.NewNopLogger(), 5)
	f := m(func(ctx context.Context, request interface{}) (response interface{}, err error) {
		span := opentracing.SpanFromContext(ctx)
		span.SetBaggageItem("foo", "bar")
		done <- struct{}{}
		return nil, nil
	})

	f(ctx, nil)
	<-done
	span.Finish()
	assert.Equal(t, "bar", tracer.FinishedSpans()[0].BaggageItem("foo"))

}
