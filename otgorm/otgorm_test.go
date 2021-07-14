package otgorm

import (
	"context"
	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

type mockModel struct {
	gorm.Model
	value string
}

func TestHook(t *testing.T) {
	var interceptorCalled bool
	tracer := mocktracer.New()
	factory, cleanup := provideDBFactory(databaseIn{
		Conf: config.MapAdapter{"gorm": map[string]databaseConf{
			"default": {
				Database: "sqlite",
				Dsn:      ":memory:",
			},
		}},
		Logger: log.NewNopLogger(),
		GormConfigInterceptor: func(name string, conf *gorm.Config) {
			interceptorCalled = true
		},
		Tracer: tracer,
	})
	defer cleanup()

	db, err := factory.Make("default")
	assert.NoError(t, err)

	_, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "test")
	db.WithContext(ctx).AutoMigrate(&mockModel{})
	assert.NotEmpty(t, tracer.FinishedSpans())

	assert.True(t, interceptorCalled)

}
