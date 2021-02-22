package queue

import (
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/events"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvideDispatcher(t *testing.T) {
	out, err := Provide(DispatcherIn{
		Conf: config.MapAdapter{"queue": map[string]configuration{
			"default": {
				1,
				5,
			},
			"alternative": {
				3,
				5,
			},
		}},
		Dispatcher:  &events.SyncDispatcher{},
		RedisClient: redis.NewUniversalClient(&redis.UniversalOptions{}),
		Logger:      log.NewNopLogger(),
		AppName:     config.AppName("test"),
		Env:         config.NewEnv("testing"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, out.QueueableDispatcher)
	assert.NotNil(t, out.DispatcherFactory)
	assert.NotNil(t, out.Dispatcher)
	assert.NotNil(t, out.DispatcherMaker)
	def, err := out.DispatcherMaker.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	assert.Implements(t, (*di.Module)(nil), out)
}
