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
	factory, cleanup, err := NewRedisFactory(RedisParam{
		In: dig.In{},
		Conf: config.MapAdapter{"redis": map[string]redis.UniversalOptions{
			"default":     {},
			"alternative": {},
		}},
		Logger: log.NewNopLogger(),
		Tracer: nil,
	})
	assert.NoError(t, err)
	assert.Len(t, factory.db, 2)
	assert.NotNil(t, cleanup)
}
