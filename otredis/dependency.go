package otredis

import (
	"fmt"
	"time"

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
		*collector
*/
func Providers() []interface{} {
	return []interface{}{provideRedisFactory, provideDefaultClient, provideConfig}
}

// RedisConfigurationInterceptor intercepts the redis.UniversalOptions before
// creating the client so you can make amendment to it. Useful because some
// configuration can not be mapped to a text representation. For example, you
// cannot add OnConnect callback factoryIn a configuration file, but you can add it
// here.
type RedisConfigurationInterceptor func(name string, opts *redis.UniversalOptions)

// factoryIn is the injection parameter for provideRedisFactory.
type factoryIn struct {
	di.In

	Logger      log.Logger
	Conf        contract.ConfigAccessor
	Interceptor RedisConfigurationInterceptor `optional:"true"`
	Tracer      opentracing.Tracer            `optional:"true"`
	Gauges      *Gauges                       `optional:"true"`
	Dispatcher  contract.Dispatcher           `optional:"true"`
}

// factoryOut is the result of provideRedisFactory.
type factoryOut struct {
	di.Out

	Maker     Maker
	Factory   Factory
	Collector *collector
}

// provideRedisFactory creates Factory and redis.UniversalClient. It is a valid
// dependency for package core.
func provideRedisFactory(p factoryIn) (factoryOut, func()) {
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			base RedisUniversalOptions
			full redis.UniversalOptions
		)
		if err := p.Conf.Unmarshal(fmt.Sprintf("redis.%s", name), &base); err != nil {
			return di.Pair{}, fmt.Errorf("redis configuration %s not valid: %w", name, err)
		}
		if len(base.Addrs) == 0 {
			base = RedisUniversalOptions{
				Addrs: []string{"127.0.0.1:6379"},
			}
		}

		full = redis.UniversalOptions{
			Addrs:              base.Addrs,
			DB:                 base.DB,
			Username:           base.Username,
			Password:           base.Password,
			SentinelPassword:   base.SentinelPassword,
			MaxRetries:         base.MaxRetries,
			MinRetryBackoff:    base.MinRetryBackoff.Duration,
			MaxRetryBackoff:    base.MaxRetryBackoff.Duration,
			DialTimeout:        base.DialTimeout.Duration,
			ReadTimeout:        base.ReadTimeout.Duration,
			WriteTimeout:       base.WriteTimeout.Duration,
			PoolSize:           base.PoolSize,
			MinIdleConns:       base.MinIdleConns,
			MaxConnAge:         base.MaxConnAge.Duration,
			PoolTimeout:        base.PoolTimeout.Duration,
			IdleTimeout:        base.IdleTimeout.Duration,
			IdleCheckFrequency: base.IdleCheckFrequency.Duration,
			TLSConfig:          nil,
			MaxRedirects:       base.MaxRetries,
			ReadOnly:           base.ReadOnly,
			RouteByLatency:     base.RouteByLatency,
			RouteRandomly:      base.RouteRandomly,
			MasterName:         base.MasterName,
		}
		if p.Interceptor != nil {
			p.Interceptor(name, &full)
		}
		redis.SetLogger(&RedisLogAdapter{level.Debug(p.Logger)})

		client := redis.NewUniversalClient(&full)
		if p.Tracer != nil {
			client.AddHook(
				hook{
					addrs:    full.Addrs,
					database: full.DB,
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
	redisFactory.SubscribeReloadEventFrom(p.Dispatcher)
	var collector *collector
	if p.Gauges != nil {
		var interval time.Duration
		p.Conf.Unmarshal("redisMetrics.interval", &interval)
		collector = newCollector(redisFactory, p.Gauges, interval)
	}
	redisOut := factoryOut{
		Maker:     redisFactory,
		Factory:   redisFactory,
		Collector: collector,
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

type metricsConf struct {
	Interval config.Duration `json:"interval" yaml:"interval"`
}

// provideConfig exports the default redis configuration
func provideConfig() configOut {
	configs := []config.ExportedConfig{
		{
			Owner: "otredis",
			Data: map[string]interface{}{
				"redis": map[string]RedisUniversalOptions{
					"default": {
						Addrs: []string{"127.0.0.1:6379"},
					},
				},
				"redisMetrics": metricsConf{
					Interval: config.Duration{Duration: 15 * time.Second},
				},
			},
			Comment: "The configuration of redis clients",
		},
	}
	return configOut{Config: configs}
}
