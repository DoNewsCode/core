package otes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/contract/lifecycle"
	"github.com/DoNewsCode/core/di"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/olivere/elastic/v7"
	"github.com/opentracing/opentracing-go"
)

/*
Providers returns a set of dependency providers. It includes the Maker, the
default *elastic.Client and exported configs.
	Depends On:
		log.Logger
		contract.ConfigAccessor
		opentracing.Tracer     `optional:"true"`
		contract.Dispatcher    `optional:"true"`
		contract.DIPopulator
	Provides:
		Factory
		Maker
		*elastic.Client
*/
func Providers(opts ...ProvidersOptionFunc) di.Deps {
	options := providersOption{
		interceptor:       func(name string, opt *Config) {},
		clientConstructor: newClient,
	}
	for _, f := range opts {
		f(&options)
	}
	return di.Deps{
		provideEsFactory(&options),
		provideDefaultClient,
		provideConfig,
		di.Bind(new(*Factory), new(Maker)),
	}
}

// EsConfigInterceptor is an injector type hint that allows user to do
// last minute modification to es configurations.
type EsConfigInterceptor func(name string, opt *Config)

// factoryIn is the injection parameter for Provide.
type factoryIn struct {
	di.In

	Logger     log.Logger
	Conf       contract.ConfigUnmarshaler
	Dispatcher lifecycle.ConfigReload `optional:"true"`
	Populator  contract.DIPopulator
}

// Provide creates Factory and *elastic.Client. It is a valid dependency for
// package core.
func provideEsFactory(option *providersOption) func(p factoryIn) (*Factory, func()) {
	if option.interceptor == nil {
		option.interceptor = func(name string, opt *Config) {}
	}
	if option.clientConstructor == nil {
		option.clientConstructor = newClient
	}
	return func(p factoryIn) (*Factory, func()) {
		factory := di.NewFactory[*elastic.Client](func(name string) (pair di.Pair[*elastic.Client], err error) {
			var conf Config
			if err := p.Conf.Unmarshal(fmt.Sprintf("es.%s", name), &conf); err != nil {
				if name != "default" {
					return pair, fmt.Errorf("elastic configuration %s not valid: %w", name, err)
				}
				conf.URL = []string{"http://127.0.0.1:9200"}
			}

			option.interceptor(name, &conf)

			client, err := option.clientConstructor(ClientArgs{
				Name:      name,
				Conf:      &conf,
				Populator: p.Populator,
			})
			if err != nil {
				return pair, err
			}

			return di.Pair[*elastic.Client]{
				Conn: client,
				Closer: func() {
					client.Stop()
				},
			}, nil
		})
		if option.reloadable && p.Dispatcher != nil {
			p.Dispatcher.On(func(ctx context.Context, Config contract.ConfigUnmarshaler) error {
				factory.Close()
				return nil
			})
		}
		return factory, factory.Close
	}
}

// ClientArgs are arguments for constructing elasticsearch clients.
// Use this as input when providing custom constructor.
type ClientArgs struct {
	Name      string
	Conf      *Config
	Populator contract.DIPopulator
}

func newClient(args ClientArgs) (*elastic.Client, error) {
	var options []elastic.ClientOptionFunc

	var injected struct {
		di.In

		opentracing.Tracer `optional:"true"`
		log.Logger
	}

	args.Populator.Populate(&injected)

	if injected.Tracer != nil {
		options = append(options,
			elastic.SetHttpClient(
				&http.Client{
					Transport: NewTransport(WithTracer(injected.Tracer)),
				},
			),
		)
	}

	if args.Conf.Healthcheck != nil {
		options = append(options, elastic.SetHealthcheck(*args.Conf.Healthcheck))
	}

	if args.Conf.Sniff != nil {
		options = append(options, elastic.SetSniff(*args.Conf.Sniff))
	}
	logger := log.With(injected.Logger, "tag", "es")
	options = append(options,
		elastic.SetURL(args.Conf.URL...),
		elastic.SetBasicAuth(args.Conf.Username, args.Conf.Password),
		elastic.SetInfoLog(ElasticLogAdapter{level.Info(logger)}),
		elastic.SetErrorLog(ElasticLogAdapter{level.Error(logger)}),
		elastic.SetTraceLog(ElasticLogAdapter{level.Debug(logger)}),
	)
	client, err := elastic.NewClient(options...)
	if err != nil {
		client, err = elastic.NewSimpleClient(options...)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func provideDefaultClient(maker Maker) (*elastic.Client, error) {
	return maker.Make("default")
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

// provideConfig exports the default es configuration.
func provideConfig() configOut {
	configs := []config.ExportedConfig{
		{
			Owner: "otes",
			Data: map[string]any{
				"es": map[string]Config{
					"default": {
						URL:    []string{"http://127.0.0.1:9200"},
						Shards: 1,
					},
				},
			},
			Comment: "The configuration of elastic search clients",
		},
	}
	return configOut{Config: configs}
}
