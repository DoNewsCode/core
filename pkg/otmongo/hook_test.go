package otmongo

import (
	"context"
	"github.com/DoNewsCode/std/pkg/core"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"testing"
)

func TestHook(t *testing.T) {
	t.Parallel()
	c := core.New()
	tracer := mocktracer.New()
	c.ProvideEssentials()
	c.Provide(func() opentracing.Tracer {
		return tracer
	})
	c.Provide(ProvideMongo)
	c.Invoke(func(mongo *mongo.Client) {
		mongo.Ping(context.Background(), readpref.Nearest())
		assert.NotEmpty(t, tracer.FinishedSpans())
	})
}
