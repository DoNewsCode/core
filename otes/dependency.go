package otes

import (
	"fmt"
	"net/http"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/internal"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/olivere/elastic/v7"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
)

var envDefaultElasticsearchAddrs, envDefaultElasticsearchAddrsIsSet = internal.GetDefaultAddrsFromEnv("ELASTICSEARCH_ADDR", "http://127.0.0.1:9200")

/*
Providers returns a set of dependency providers. It includes the Maker, the
default *elastic.Client and exported configs.
	Depends On:
		log.Logger
		contract.ConfigAccessor
		EsConfigInterceptor `optional:"true"`
		opentracing.Tracer     `optional:"true"`
	Provides:
		Factory
		Maker
		*elastic.Client
*/
func Providers() di.Deps {
	return []interface{}{provideEsFactory, provideDefaultClient, provideConfig}
}

// EsConfigInterceptor is an injector type hint that allows user to do
// last minute modification to es configurations.
type EsConfigInterceptor func(name string, opt *Config)

// factoryIn is the injection parameter for Provide.
type factoryIn struct {
	dig.In

	Logger      log.Logger
	Conf        contract.ConfigAccessor
	Interceptor EsConfigInterceptor        `optional:"true"`
	Tracer      opentracing.Tracer         `optional:"true"`
	Options     []elastic.ClientOptionFunc `optional:"true"`
	Dispatcher  contract.Dispatcher        `optional:"true"`
}

// factoryOut is the result of Provide.
type factoryOut struct {
	dig.Out

	Factory        Factory
	Maker          Maker
	ExportedConfig []config.ExportedConfig `group:"config,flatten"`
}

// Provide creates Factory and *elastic.Client. It is a valid dependency for
// package core.
func provideEsFactory(p factoryIn) (factoryOut, func()) {
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			conf    Config
			options []elastic.ClientOptionFunc
		)
		if err := p.Conf.Unmarshal(fmt.Sprintf("es.%s", name), &conf); err != nil {
			if name != "default" {
				return di.Pair{}, fmt.Errorf("elastic configuration %s not valid: %w", name, err)
			}
			conf.URL = envDefaultElasticsearchAddrs
		}
		if p.Interceptor != nil {
			p.Interceptor(name, &conf)
		}

		if p.Tracer != nil {
			options = append(options,
				elastic.SetHttpClient(
					&http.Client{
						Transport: NewTransport(WithTracer(p.Tracer)),
					},
				),
			)
		}

		if conf.Healthcheck != nil {
			options = append(options, elastic.SetHealthcheck(*conf.Healthcheck))
		}

		if conf.Sniff != nil {
			options = append(options, elastic.SetSniff(*conf.Sniff))
		}
		logger := log.With(p.Logger, "tag", "es")
		options = append(options,
			elastic.SetURL(conf.URL...),
			elastic.SetBasicAuth(conf.Username, conf.Password),
			elastic.SetInfoLog(ElasticLogAdapter{level.Info(logger)}),
			elastic.SetErrorLog(ElasticLogAdapter{level.Error(logger)}),
			elastic.SetTraceLog(ElasticLogAdapter{level.Debug(logger)}),
		)
		options = append(options, p.Options...)

		client, err := elastic.NewClient(options...)
		if err != nil {
			return di.Pair{}, err
		}

		return di.Pair{
			Conn: client,
			Closer: func() {
				client.Stop()
			},
		}, nil
	})
	f := Factory{factory}
	f.SubscribeReloadEventFrom(p.Dispatcher)
	return factoryOut{
		Factory: f,
		Maker:   f,
	}, factory.Close
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
			Data: map[string]interface{}{
				"es": map[string]Config{
					"default": {
						URL:    envDefaultElasticsearchAddrs,
						Shards: 1,
					},
				},
			},
			Comment: "The configuration of elastic search clients",
		},
	}
	return configOut{Config: configs}
}
