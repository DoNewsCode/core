package otfranz

import (
	"fmt"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/opentracing/opentracing-go"
	"github.com/twmb/franz-go/pkg/kgo"
)

/*
Providers is a set of dependencies including Factory,Maker and exported configs.
	Depends On:
		opentracing.Tracer
		contract.ConfigUnmarshaler
		log.Logger
		contract.Dispatcher
	Provide:
		Factory
		Maker
*/
func Providers(optionFunc ...ProvidersOptionFunc) di.Deps {
	option := providersOption{
		interceptor: func(name string, conf *Config) {},
	}
	for _, f := range optionFunc {
		f(&option)
	}
	return di.Deps{
		provideKafkaFactory(&option),
		provideConfig,
		di.Bind(new(Factory), new(Maker)),
	}
}

// Maker models a Factory
type Maker interface {
	Make(name string) (*kgo.Client, error)
}

// factoryIn is an injection parameter for provideFactory.
type factoryIn struct {
	di.In

	Tracer     opentracing.Tracer `optional:"true"`
	Conf       contract.ConfigUnmarshaler
	Logger     log.Logger
	Dispatcher contract.Dispatcher `optional:"true"`
}

// factoryOut is the result of provideKafkaFactory.
type factoryOut struct {
	di.Out

	Factory Factory
	Client  *kgo.Client
}

// Module implements di.Modular
func (f factoryOut) Module() interface{} {
	return f
}

// provideKafkaFactory creates the Factory. It is valid dependency option for package core.
// Note: when working with package core's DI container, use provideKafkaFactory over provideFactory.
// Not only provideKafkaFactory provides *kgo.Client, but also only provideKafkaFactory
// exports default Kafka configuration.
func provideKafkaFactory(option *providersOption) func(p factoryIn) (factoryOut, func(), error) {
	if option.interceptor == nil {
		option.interceptor = func(name string, config *Config) {}
	}

	return func(p factoryIn) (factoryOut, func(), error) {
		factory, rc := provideFactory(p, option.interceptor)
		cli, err := factory.Make("default")
		if err != nil {
			level.Warn(p.Logger).Log("err", err)
		}

		if p.Dispatcher != nil {
			if option.reloadable {
				factory.SubscribeReloadEventFrom(p.Dispatcher)
			}
		}

		return factoryOut{
			Factory: factory,
			Client:  cli,
		}, rc, nil
	}
}

// provideFactory creates the Factory. It is valid dependency option for package core.
func provideFactory(p factoryIn, interceptor Interceptor) (Factory, func()) {
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			err  error
			conf Config
		)

		err = p.Conf.Unmarshal(fmt.Sprintf("kafka.%s", name), &conf)
		if err != nil {
			return di.Pair{}, fmt.Errorf("kafka configuration %s not valid: %w", name, err)
		}
		if p.Logger != nil {
			var logLevel string
			err = p.Conf.Unmarshal("log.level", &logLevel)
			if err != nil {
				logLevel = "debug"
			}
			conf.Logger = FranzLogAdapter(logLevel, p.Logger)
		}

		interceptor(name, &conf)
		// converts Config to []kgo.Opt
		opts := fromConfig(conf)

		client, err := kgo.NewClient(opts...)
		if err != nil {
			return di.Pair{}, err
		}
		return di.Pair{
			Conn: client,
			Closer: func() {
				client.Close()
			},
		}, nil
	})
	return Factory{factory}, factory.Close
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

func provideConfig() configOut {
	configs := []config.ExportedConfig{
		{
			Owner: "otfranz",
			Data: map[string]interface{}{
				"kafka": map[string]interface{}{
					"default": Config{
						SeedBrokers:         []string{"127.0.0.1:9092"},
						DefaultProduceTopic: "example",
						Topics:              []string{"example"},
						Group:               "example",
					},
				},
			},
			Comment: "The configuration of kafka with franz-go",
		},
	}
	return configOut{Config: configs}
}
