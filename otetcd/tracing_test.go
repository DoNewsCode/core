package otetcd

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"

	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

func TestTracing(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("Set env ETCD_ADDR to run TestTracing")
		return
	}
	addrs := strings.Split(os.Getenv("ETCD_ADDR"), ",")
	var interceptorCalled bool
	tracer := mocktracer.New()
	factory, cleanup := provideFactory(&providersOption{
		interceptor: func(name string, options *clientv3.Config) {
			interceptorCalled = true
			assert.Equal(t, "default", name)
		},
	})(factoryIn{
		Logger: log.NewNopLogger(),
		Conf: config.MapAdapter{"etcd": map[string]Option{
			"default": {
				Endpoints: addrs,
			},
		}},
		Tracer: tracer,
	})
	defer cleanup()

	client, err := factory.Make("default")
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "test")
	defer span.Finish()

	response, err := client.Get(ctx, "foo")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, tracer.FinishedSpans())
	assert.True(t, interceptorCalled)
}
