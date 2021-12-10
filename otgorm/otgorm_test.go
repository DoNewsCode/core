package otgorm

import (
	"context"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockModel struct {
	gorm.Model
}

func TestHook(t *testing.T) {
	var interceptorCalled bool
	tracer := mocktracer.New()
	out, cleanup, _ := provideDBFactory(&providersOption{
		interceptor: func(name string, conf *gorm.Config) {
			interceptorCalled = true
		},
		drivers: map[string]func(dsn string) gorm.Dialector{"sqlite": sqlite.Open},
	})(factoryIn{
		Conf: config.MapAdapter{"gorm": map[string]databaseConf{
			"default": {
				Database: "sqlite",
				Dsn:      ":memory:",
			},
		}},
		Logger: log.NewNopLogger(),
		Tracer: tracer,
	})
	defer cleanup()

	factory := out.Factory

	db, err := factory.Make("default")
	assert.NoError(t, err)

	_, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "test")
	db.WithContext(ctx).AutoMigrate(&mockModel{})
	assert.NotEmpty(t, tracer.FinishedSpans())

	assert.True(t, interceptorCalled)
}
