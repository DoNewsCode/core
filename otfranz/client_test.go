package otfranz

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
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestClient_ProduceWithTracing(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestProvideFactory")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	factory, cleanup := provideFactory(factoryIn{
		Logger: log.NewNopLogger(),
		Conf: config.MapAdapter{"kafka": map[string]Config{
			"default": {
				SeedBrokers:         addrs,
				DefaultProduceTopic: franzTestTopic,
			},
		}},
	}, func(name string, config *Config) {})
	defer cleanup()
	cli, err := factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, cli)

	tracer := mocktracer.New()

	clientWithTrace := NewClient(cli, tracer)

	span, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "test")
	record := &kgo.Record{Value: []byte("bar")}
	clientWithTrace.ProduceWithTracing(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			t.Fatalf("produce error: %v\n", err)
		}
	})
	time.Sleep(time.Second)

	if err := clientWithTrace.ProduceSyncWithTracing(ctx, record).FirstErr(); err != nil {
		t.Fatalf("produce sync error: %v\n", err)
	}

	assert.Len(t, tracer.FinishedSpans(), 2)
	span.Finish()

}
