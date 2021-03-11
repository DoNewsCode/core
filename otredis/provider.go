package otredis

import (
	"fmt"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
)

/*
Providers returns a set of dependency providers related to redis. It includes the
Maker, the default redis.UniversalClient and exported configs.

	Depends On:
		log.Logger
		contract.ConfigAccessor
		RedisConfigurationInterceptor `optional:"true"`
		opentracing.Tracer            `optional:"true"`
	Provide:
		Maker
		Factory
		redis.UniversalClient
*/
func Providers() []interface{} {
	return []interface{}{provideRedisFactory, provideDefaultClient, provideConfig}
}

// RedisConfigurationInterceptor intercepts the redis.UniversalOptions before
// creating the client so you can make amendment to it. Useful because some
// configuration can not be mapped to a text representation. For example, you
// cannot add OnConnect callback in a configuration file, but you can add it
// here.
type RedisConfigurationInterceptor func(name string, opts *redis.UniversalOptions)

// Maker is models Factory
type Maker interface {
	Make(name string) (redis.UniversalClient, error)
}

// Factory is a *di.Factory that creates redis.UniversalClient using a
// specific configuration entry.
type Factory struct {
	*di.Factory
}

// Make creates redis.UniversalClient using a specific configuration entry.
func (r Factory) Make(name string) (redis.UniversalClient, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(redis.UniversalClient), nil
}

// in is the injection parameter for provideRedisFactory.
type in struct {
	di.In

	Logger      log.Logger
	Conf        contract.ConfigAccessor
	Interceptor RedisConfigurationInterceptor `optional:"true"`
	Tracer      opentracing.Tracer            `optional:"true"`
}

// out is the result of provideRedisFactory.
type out struct {
	di.Out

	Maker   Maker
	Factory Factory
}

// provideRedisFactory creates Factory and redis.UniversalClient. It is a valid
// dependency for package core.
func provideRedisFactory(p in) (out, func()) {
	var err error
	var dbConfs map[string]redis.UniversalOptions
	err = p.Conf.Unmarshal("redis", &dbConfs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			ok   bool
			conf redis.UniversalOptions
		)
		if conf, ok = dbConfs[name]; !ok {
			if name != "default" {
				return di.Pair{}, fmt.Errorf("redis configuration %s not valid", name)
			}
			conf = redis.UniversalOptions{}
		}
		if p.Interceptor != nil {
			p.Interceptor(name, &conf)
		}
		client := redis.NewUniversalClient(&conf)
		if p.Logger != nil {
			redis.SetLogger(&RedisLogAdapter{level.Debug(p.Logger)})
		}
		if p.Tracer != nil {
			client.AddHook(
				hook{
					addrs:    conf.Addrs,
					database: conf.DB,
					tracer:   p.Tracer,
				},
			)
		}
		return di.Pair{
			Conn: client,
			Closer: func() {
				_ = client.Close()
			},
		}, nil
	})
	redisFactory := Factory{factory}
	redisOut := out{
		Maker:   redisFactory,
		Factory: redisFactory,
	}
	return redisOut, redisFactory.Close
}

func provideDefaultClient(maker Maker) (redis.UniversalClient, error) {
	return maker.Make("default")
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

// provideConfig exports the default redis configuration
func provideConfig() configOut {
	configs := []config.ExportedConfig{
		{
			Owner: "otredis",
			Data: map[string]interface{}{
				"redis": map[string]map[string]interface{}{
					"default": {
						"addrs":              []string{"127.0.0.1:6379"},
						"DB":                 0,
						"username":           "",
						"password":           "",
						"sentinelPassword":   "",
						"maxRetries":         0,
						"minRetryBackoff":    0,
						"maxRetryBackoff":    0,
						"dialTimeout":        0,
						"readTimeout":        0,
						"writeTimeout":       0,
						"poolSize":           0,
						"minIdleConns":       0,
						"maxConnAge":         0,
						"poolTimeout":        0,
						"idleTimeout":        0,
						"idleCheckFrequency": 0,
						"maxRedirects":       0,
						"readOnly":           false,
						"routeByLatency":     false,
						"routeRandomly":      false,
						"masterName":         "",
					},
				},
			},
			Comment: "The configuration of redis clients",
		},
	}
	return configOut{Config: configs}
}
