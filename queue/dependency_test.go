package queue

import (
	"context"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/events"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

type maker struct{}

func (m maker) Make(name string) (redis.UniversalClient, error) {
	return redis.NewUniversalClient(&redis.UniversalOptions{Addrs: []string{os.Getenv("REDIS_ADDR")}}), nil
}

func TestProvideDispatcher(t *testing.T) {
	out, err := provideDispatcherFactory(makerIn{
		Conf: config.MapAdapter{"queue": map[string]configuration{
			"default": {
				"default",
				1,
				5,
			},
			"alternative": {
				"default",
				3,
				5,
			},
		}},
		Dispatcher: &events.SyncDispatcher{},
		RedisMaker: maker{},
		Logger:     log.NewNopLogger(),
		AppName:    config.AppName("test"),
		Env:        config.EnvTesting,
	})
	assert.NoError(t, err)
	assert.NotNil(t, out.DispatcherFactory)
	assert.NotNil(t, out.DispatcherMaker)
	def, err := out.DispatcherMaker.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.Implements(t, (*di.Module)(nil), out)
}

type mockDriver struct {
}

func (m mockDriver) Push(ctx context.Context, message *PersistedEvent, delay time.Duration) error {
	panic("implement me")
}

func (m mockDriver) Pop(ctx context.Context) (*PersistedEvent, error) {
	panic("implement me")
}

func (m mockDriver) Ack(ctx context.Context, message *PersistedEvent) error {
	panic("implement me")
}

func (m mockDriver) Fail(ctx context.Context, message *PersistedEvent) error {
	panic("implement me")
}

func (m mockDriver) Reload(ctx context.Context, channel string) (int64, error) {
	panic("implement me")
}

func (m mockDriver) Flush(ctx context.Context, channel string) error {
	panic("implement me")
}

func (m mockDriver) Info(ctx context.Context) (QueueInfo, error) {
	panic("implement me")
}

func (m mockDriver) Retry(ctx context.Context, message *PersistedEvent) error {
	panic("implement me")
}

func TestProvideDispatcher_withDriver(t *testing.T) {
	out, err := provideDispatcherFactory(makerIn{
		Conf: config.MapAdapter{"queue": map[string]configuration{
			"default": {
				"default",
				1,
				5,
			},
			"alternative": {
				"default",
				3,
				5,
			},
		}},
		Dispatcher: &events.SyncDispatcher{},
		Driver:     mockDriver{},
		Logger:     log.NewNopLogger(),
		AppName:    config.AppName("test"),
		Env:        config.EnvTesting,
	})
	assert.NoError(t, err)
	assert.NotNil(t, out.DispatcherFactory)
	assert.NotNil(t, out.DispatcherMaker)
	def, err := out.DispatcherMaker.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.Implements(t, (*di.Module)(nil), out)
}

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}
