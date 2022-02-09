package otgorm

import (
	"context"
	"fmt"
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
		Conf: config.MapAdapter{
			"gorm": map[string]interface{}{
				"default": map[string]interface{}{
					"database": "sqlite",
					"dsn":      ":memory:",
				},
			},
		},
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

func TestHook_raw(t *testing.T) {
	tracer := mocktracer.New()
	out, cleanup, _ := provideDBFactory(&providersOption{
		drivers: map[string]func(dsn string) gorm.Dialector{"sqlite": sqlite.Open},
	})(factoryIn{
		Conf: config.MapAdapter{
			"gorm": map[string]interface{}{
				"default": map[string]interface{}{
					"database": "sqlite",
					"dsn":      ":memory:",
				},
			},
		},
		Logger: log.NewNopLogger(),
		Tracer: tracer,
	})
	defer cleanup()

	factory := out.Factory

	db, err := factory.Make("default")
	assert.NoError(t, err)

	_, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "test")

	err = db.WithContext(ctx).Exec("CREATE TABLE test (id uint)").Error
	assert.NoError(t, err)

	err = db.WithContext(ctx).Exec("INSERT INTO test (id) VALUES (1)").Error
	assert.NoError(t, err)

	err = db.WithContext(ctx).Exec("INSERT INTO test (id) VALUES (2)").Error
	assert.NoError(t, err)

	rows, err := db.WithContext(ctx).Raw("SELECT * FROM test").Rows()
	assert.NoError(t, err)

	var models []mockModel
	for rows.Next() {
		fmt.Println("next")
		var m mockModel
		err = db.WithContext(ctx).ScanRows(rows, &m)
		assert.NoError(t, err)
		models = append(models, m)
	}
	t.Log(models)

	db.WithContext(ctx).Raw("SELECT * FROM test").Scan(&models)
	t.Log(models)

	assert.Len(t, tracer.FinishedSpans(), 5)
}
