package otredis

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisFactory(t *testing.T) {
	redisOut, cleanup := provideRedisFactory(in{
		Conf: config.MapAdapter{"redis": map[string]redis.UniversalOptions{
			"default":     {},
			"alternative": {},
		}},
		Logger: log.NewNopLogger(),
		Tracer: nil,
	})
	def, err := redisOut.Maker.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := redisOut.Maker.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	assert.NotNil(t, cleanup)
	cleanup()
}
