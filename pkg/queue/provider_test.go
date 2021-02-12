package queue

import (
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/core"
	"github.com/DoNewsCode/std/pkg/event"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvideDispatcher(t *testing.T) {
	out, err := ProvideDispatcher(DispatcherIn{
		Conf: config.MapAdapter{"queue": map[string]queueConf{
			"default": {
				1,
				5,
			},
			"alternative": {
				3,
				5,
			},
		}},
		Dispatcher:  &event.SyncDispatcher{},
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
	assert.Implements(t, (*core.Module)(nil), out)
}
