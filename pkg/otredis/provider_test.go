package otredis

import (
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
	"testing"
)

func TestNewRedisFactory(t *testing.T) {
	factory, cleanup := ProvideRedisFactory(RedisParam{
		In: dig.In{},
		Conf: config.MapAdapter{"redis": map[string]redis.UniversalOptions{
			"default":     {},
			"alternative": {},
		}},
		Logger: log.NewNopLogger(),
		Tracer: nil,
	})
	def, err := factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := factory.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	assert.NotNil(t, cleanup)
	cleanup()
}
