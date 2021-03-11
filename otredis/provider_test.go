package otredis

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	yaml2 "github.com/ghodss/yaml"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
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

func TestProvideConfigs(t *testing.T) {
	var r redis.UniversalOptions
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
	c.Config[0].Data["redis"].(map[string]map[string]interface{})["default"]["db"] = 1
	c.Config[0].Data["redis"].(map[string]map[string]interface{})["default"]["addrs"] = []string{"0.0.0.0:6379"}
	bytes, _ := yaml2.Marshal(c.Config[0].Data)
	k := koanf.New(".")
	k.Load(rawbytes.Provider(bytes), yaml.Parser())
	k.Unmarshal("redis.default", &r)
	assert.Equal(t, 1, r.DB)
	assert.Equal(t, []string{"0.0.0.0:6379"}, r.Addrs)
}
