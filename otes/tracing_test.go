package otes

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

func TestTracing(t *testing.T) {
	if os.Getenv("ELASTICSEARCH_ADDR") == "" {
		t.Skip("set env ELASTICSEARCH_ADDR to run TestTracing")
		return
	}
	addrs := strings.Split(os.Getenv("ELASTICSEARCH_ADDR"), ",")
	tracer := mocktracer.New()
	opentracing.SetGlobalTracer(tracer)
	factory, cleanup := provideEsFactory(&providersOption{})(factoryIn{
		Conf: config.MapAdapter{"es": map[string]Config{
			"default":     {URL: addrs},
			"alternative": {URL: addrs},
		}},
		Logger:    log.NewNopLogger(),
		Populator: Populator{},
	})
	defer cleanup()

	client, err := factory.Make("default")
	assert.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "es.query")
	defer span.Finish()

	res, code, err := client.Ping(addrs[0]).Do(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NotNil(t, res)
	assert.NotEmpty(t, tracer.FinishedSpans())
}
