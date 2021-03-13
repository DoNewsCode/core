package otes

import (
	"fmt"
	"net/http"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/olivere/elastic/v7"
	esConfig "github.com/olivere/elastic/v7/config"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
)

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
type EsConfigInterceptor func(name string, opt *esConfig.Config)

// Maker models Factory
type Maker interface {
	Make(name string) (*elastic.Client, error)
}

// Factory is a *di.Factory that creates *elastic.Client using a specific
// configuration entry.
type Factory struct {
	*di.Factory
}

// Make creates *elastic.Client using a specific configuration entry.
func (r Factory) Make(name string) (*elastic.Client, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*elastic.Client), nil
}

// in is the injection parameter for Provide.
type in struct {
	dig.In

	Logger      log.Logger
	Conf        contract.ConfigAccessor
	Interceptor EsConfigInterceptor `optional:"true"`
	Tracer      opentracing.Tracer  `optional:"true"`
}

// out is the result of Provide.
type out struct {
	dig.Out

	Factory        Factory
	Maker          Maker
	ExportedConfig []config.ExportedConfig `group:"config,flatten"`
}

// Provide creates Factory and *elastic.Client. It is a valid dependency for
// package core.
func provideEsFactory(p in) (out, func()) {
	var err error
	var esConfigs map[string]esConfig.Config
	err = p.Conf.Unmarshal("es", &esConfigs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			ok      bool
			conf    esConfig.Config
			options []elastic.ClientOptionFunc
		)
		if conf, ok = esConfigs[name]; !ok {
			if name != "default" {
				return di.Pair{}, fmt.Errorf("elastic configuration %s not valid", name)
			}
			conf.URL = "http://localhost:9200"
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
			elastic.SetURL(conf.URL),
			elastic.SetBasicAuth(conf.Username, conf.Password),
			elastic.SetInfoLog(esLogAdapter{level.Info(logger)}),
			elastic.SetErrorLog(esLogAdapter{level.Error(logger)}),
			elastic.SetTraceLog(esLogAdapter{level.Debug(logger)}),
		)

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
	return out{
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
				"es": map[string]map[string]interface{}{
					"default": {
						"url":         "http://localhost:9200",
						"index":       "",
						"username":    "",
						"password":    "",
						"shards":      1,
						"replicas":    0,
						"sniff":       false,
						"healthCheck": false,
					},
				},
			},
			Comment: "The configuration of elastic search clients",
		},
	}
	return configOut{Config: configs}
}
