package otredis

import (
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
	"sync"
)

type RedisParam struct {
	dig.In

	Logger log.Logger
	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

func Redis(param RedisParam) (redis.UniversalClient, func()) {
	var conf redis.UniversalOptions
	_ = param.Conf.Unmarshal("redis.default", &conf)
	client := redis.NewUniversalClient(&conf)
	if param.Tracer != nil {
		client.AddHook(
			NewHook(param.Tracer, param.Conf.Strings("redis.default.addrs"), param.Conf.Int("redis.default.database")),
		)
	}

	return client, func() {
		if err := client.Close(); err != nil {
			level.Error(param.Logger).Log("err", err.Error())
		}
	}
}

type RedisFactory struct {
	db map[string]redis.UniversalClient
}

func NewRedisFactory(p RedisParam) (*RedisFactory, func(), error) {
	var err error

	var dbConfs map[string]redis.UniversalOptions
	err = p.Conf.Unmarshal("redis", &dbConfs)
	if err != nil {
		return nil, nil, err
	}
	redisFactory := &RedisFactory{
		db: make(map[string]redis.UniversalClient),
	}
	for name, value := range dbConfs {
		client := redis.NewUniversalClient(&value)
		if p.Tracer != nil {
			client.AddHook(
				NewHook(p.Tracer, value.Addrs, value.DB),
			)
		}
		redisFactory.db[name] = client
	}
	return redisFactory, func() {
		var wg sync.WaitGroup
		for i := range redisFactory.db {
			wg.Add(1)
			go func(i string) {
				_ = redisFactory.db[i].Close()
				wg.Done()
			}(i)
		}
		wg.Wait()
	}, nil
}

func (r *RedisFactory) Connection(name string) redis.UniversalClient {
	return r.db[name]
}
