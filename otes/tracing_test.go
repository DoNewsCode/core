// +build integration

package otes

import (
	"context"
	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	esConfig "github.com/olivere/elastic/v7/config"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestTracing(t *testing.T) {
	tracer := mocktracer.New()
	opentracing.SetGlobalTracer(tracer)
	factory, cleanup := provideEsFactory(in{
		Conf: config.MapAdapter{"es": map[string]esConfig.Config{
			"default":     {URL: "http://localhost:9200"},
			"alternative": {URL: "http://localhost:9200"},
		}},
		Logger: log.NewNopLogger(),
		Tracer: tracer,
	})
	defer cleanup()

	client, err := factory.Maker.Make("default")
	assert.NoError(t, err)
	span, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "es.query")
	defer span.Finish()

	res, code, err := client.Ping("http://localhost:9200").Do(ctx)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
	assert.NotNil(t, res)
	assert.NotEmpty(t, tracer.FinishedSpans())
}
