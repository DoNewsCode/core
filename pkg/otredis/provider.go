package otredis

import (
	"fmt"
	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
)

type RedisParam struct {
	dig.In

	Logger log.Logger
	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

func ProvideDefaultRedis(p RedisParam) (redis.UniversalClient, func(), error) {
	factory, _ := ProvideRedisFactory(p)
	conn, err := factory.Make("default")
	return conn, func() {
		factory.CloseConn("default")
	}, err
}

type RedisFactory struct {
	*async.Factory
}

func (r RedisFactory) Make(name string) (redis.UniversalClient, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(redis.UniversalClient), nil
}

func ProvideRedisFactory(p RedisParam) (RedisFactory, func()) {
	var err error
	var dbConfs map[string]redis.UniversalOptions
	err = p.Conf.Unmarshal("redis", &dbConfs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok   bool
			conf redis.UniversalOptions
		)
		if conf, ok = dbConfs[name]; !ok {
			return async.Pair{}, fmt.Errorf("redis configuration %s not valid", name)
		}
		client := redis.NewUniversalClient(&conf)
		if p.Tracer != nil {
			client.AddHook(
				NewHook(p.Tracer, conf.Addrs, conf.DB),
			)
		}
		return async.Pair{
			Conn: client,
			Closer: func() {
				_ = client.Close()
			},
		}, nil
	})
	return RedisFactory{factory}, factory.Close
}
