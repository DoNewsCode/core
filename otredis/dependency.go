package otredis

import (
	"context"
	"fmt"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/contract/lifecycle"
	"github.com/DoNewsCode/core/di"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-redis/redis/v8"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
)

/*
Providers returns a set of dependency providers related to redis. It includes the
Maker, the default redis.UniversalClient and exported configs.

	Depends On:
		log.Logger
		contract.ConfigAccessor
		opentracing.Tracer            `optional:"true"`
	Provide:
		Maker
		Factory
		redis.UniversalClient
		*collector
*/
func Providers(opts ...ProvidersOptionFunc) di.Deps {
	option := providersOption{
		interceptor: func(name string, opts *redis.UniversalOptions) {},
	}
	for _, f := range opts {
		f(&option)
	}
	return di.Deps{
		provideRedisFactory(&option),
		provideDefaultClient,
		provideConfig,
		di.Bind(new(*Factory), new(Maker)),
	}
}

// factoryIn is the injection parameter for provideRedisFactory.
type factoryIn struct {
	di.In

	Logger      log.Logger
	Conf        contract.ConfigUnmarshaler
	Interceptor RedisConfigurationInterceptor `optional:"true"`
	Tracer      opentracing.Tracer            `optional:"true"`
	Gauges      *Gauges                       `optional:"true"`
	Dispatcher  lifecycle.ConfigReload        `optional:"true"`
}

// factoryOut is the result of provideRedisFactory.
type factoryOut struct {
	di.Out

	Factory   *Factory
	Collector *collector
}

// Module implements di.Module
func (m factoryOut) Module() any {
	return m
}

// ProvideRunGroup add a goroutine to periodically scan redis connections and
// report them to metrics collector such as prometheus.
func (m factoryOut) ProvideRunGroup(group *run.Group) {
	if m.Collector == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(m.Collector.interval)
	group.Add(func() error {
		for {
			select {
			case <-ticker.C:
				m.Collector.collectConnectionStats()
			case <-ctx.Done():
				ticker.Stop()
				return nil
			}
		}
	}, func(err error) {
		cancel()
	})
}

// provideRedisFactory creates Factory and redis.UniversalClient. It is a valid
// dependency for package core.
func provideRedisFactory(option *providersOption) func(p factoryIn) (factoryOut, func()) {
	if option.interceptor == nil {
		option.interceptor = func(name string, opts *redis.UniversalOptions) {}
	}
	return func(p factoryIn) (factoryOut, func()) {
		factory := di.NewFactory[redis.UniversalClient](func(name string) (pair di.Pair[redis.UniversalClient], err error) {
			var (
				base RedisUniversalOptions
				full redis.UniversalOptions
			)
			if err := p.Conf.Unmarshal(fmt.Sprintf("redis.%s", name), &base); err != nil {
				return pair, fmt.Errorf("redis configuration %s not valid: %w", name, err)
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
			option.interceptor(name, &full)
			redis.SetLogger(&RedisLogAdapter{level.Debug(log.With(p.Logger, "tag", "redis"))})

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
			return di.Pair[redis.UniversalClient]{
				Conn: client,
				Closer: func() {
					_ = client.Close()
				},
			}, nil
		})
		if option.reloadable && p.Dispatcher != nil {
			p.Dispatcher.On(func(ctx context.Context, Config contract.ConfigUnmarshaler) error {
				factory.Close()
				return nil
			})
		}
		var collector *collector
		if p.Gauges != nil {
			var interval time.Duration = 15 * time.Second
			p.Conf.Unmarshal("redisMetrics.interval", &interval)
			collector = newCollector(factory, p.Gauges, interval)
		}
		redisOut := factoryOut{
			Factory:   factory,
			Collector: collector,
		}

		return redisOut, factory.Close
	}
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
			Data: map[string]any{
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
