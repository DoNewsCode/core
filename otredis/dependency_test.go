package otredis

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/events"
	"github.com/go-kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/stretchr/testify/assert"
	yaml2 "gopkg.in/yaml.v3"
)

func TestNewRedisFactory(t *testing.T) {
	for _, c := range []struct {
		name   string
		reload bool
	}{
		{"reload", true},
		{"not reload", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := &events.SyncDispatcher{}
			redisOut, cleanup := provideRedisFactory(&providersOption{reloadable: c.reload})(factoryIn{
				Conf: config.MapAdapter{"redis": map[string]RedisUniversalOptions{
					"default":     {},
					"alternative": {},
				}},
				Logger:     log.NewNopLogger(),
				Tracer:     nil,
				Dispatcher: dispatcher,
			})
			def, err := redisOut.Factory.Make("default")
			assert.NoError(t, err)
			assert.NotNil(t, def)
			alt, err := redisOut.Factory.Make("alternative")
			assert.NoError(t, err)
			assert.NotNil(t, alt)
			assert.NotNil(t, cleanup)
			assert.Equal(t, c.reload, dispatcher.ListenerCount(events.OnReload) == 1)
			cleanup()
		})
	}
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
