package otredis

import (
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
)

func Redis(logging log.Logger, conf contract.ConfigAccessor, tracer opentracing.Tracer) (redis.UniversalClient, func()) {
	client := redis.NewUniversalClient(
		&redis.UniversalOptions{
			Addrs:    conf.Strings("redis.default.addrs"),
			DB:       conf.Int("redis.default.database"),
			Password: conf.String("redis.default.password"),
		})
	client.AddHook(
		NewHook(tracer, conf.Strings("redis.addrs"),
			conf.Int("redis.database")))
	return client, func() {
		if err := client.Close(); err != nil {
			level.Error(logging).Log("err", err.Error())
		}
	}
}
