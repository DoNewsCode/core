package otmongo

import (
	"context"
	"os"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestHook(t *testing.T) {
	if os.Getenv("MONGO_ADDR") == "" {
		t.Skip("set MONGO_ADDR to run TestHook")
		return
	}
	t.Parallel()
	c := core.New()
	tracer := mocktracer.New()
	c.ProvideEssentials()
	c.Provide(di.Deps{func() opentracing.Tracer {
		return tracer
	}})
	c.Provide(Providers())
	c.Invoke(func(mongo *mongo.Client) {
		mongo.Ping(context.Background(), readpref.Nearest())
		assert.NotEmpty(t, tracer.FinishedSpans())
	})
}
