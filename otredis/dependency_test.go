package otredis

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/stretchr/testify/assert"
	yaml2 "gopkg.in/yaml.v3"
)

func TestNewRedisFactory(t *testing.T) {
	redisOut, cleanup := provideRedisFactory(factoryIn{
		Conf: config.MapAdapter{"redis": map[string]RedisUniversalOptions{
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

func TestProvideConfigs(t *testing.T) {
	var r redis.UniversalOptions
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
	bytes, _ := yaml2.Marshal(c.Config[0].Data)
	k := koanf.New(".")
	k.Load(rawbytes.Provider(bytes), yaml.Parser())
	k.Unmarshal("redis.default", &r)
	assert.Equal(t, 0, r.DB)
	assert.Equal(t, []string{"127.0.0.1:6379"}, r.Addrs)
}
