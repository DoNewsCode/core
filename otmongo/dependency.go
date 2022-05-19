package otmongo

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/contract/lifecycle"
	"github.com/DoNewsCode/core/di"

	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/dig"
)

/*
Providers returns a set of dependency providers. It includes the Maker, the
default mongo.Client and exported configs.
	Depends On:
		log.Logger
		contract.ConfigAccessor
		MongoConfigInterceptor `optional:"true"`
		opentracing.Tracer     `optional:"true"`
	Provides:
		Factory
		Maker
		*mongo.Client
*/
func Providers(optionFunc ...ProvidersOptionFunc) di.Deps {
	o := providersOption{interceptor: func(name string, clientOptions *options.ClientOptions) {}}
	for _, f := range optionFunc {
		f(&o)
	}
	return di.Deps{
		provideMongoFactory(&o),
		provideDefaultClient,
		provideConfig,
		di.Bind(new(*Factory), new(Maker)),
	}
}

// factoryIn is the injection parameter for Provide.
type factoryIn struct {
	dig.In

	Logger     log.Logger
	Conf       contract.ConfigUnmarshaler
	Tracer     opentracing.Tracer     `optional:"true"`
	Dispatcher lifecycle.ConfigReload `optional:"true"`
}

// Provide creates Factory and *mongo.Client. It is a valid dependency for
// package core.
func provideMongoFactory(providerOption *providersOption) func(p factoryIn) (*Factory, func()) {
	if providerOption.interceptor == nil {
		providerOption.interceptor = func(name string, clientOptions *options.ClientOptions) {}
	}
	return func(p factoryIn) (*Factory, func()) {
		factory := di.NewFactory[*mongo.Client](func(name string) (pair di.Pair[*mongo.Client], err error) {
			var conf struct{ URI string }
			if err := p.Conf.Unmarshal(fmt.Sprintf("mongo.%s", name), &conf); err != nil {
				return pair, fmt.Errorf("mongo configuration %s not valid: %w", name, err)
			}
			if conf.URI == "" {
				conf.URI = "mongodb://127.0.0.1:27017"
			}

			opts := options.Client()
			opts.ApplyURI(conf.URI)
			if p.Tracer != nil {
				opts.Monitor = NewMonitor(p.Tracer)
			}
			providerOption.interceptor(name, opts)
			client, err := mongo.Connect(context.Background(), opts)
			if err != nil {
				return pair, err
			}

			return di.Pair[*mongo.Client]{
				Conn: client,
				Closer: func() {
					_ = client.Disconnect(context.Background())
				},
			}, nil
		})
		if providerOption.reloadable && p.Dispatcher != nil {
			p.Dispatcher.On(func(_ context.Context, _ contract.ConfigUnmarshaler) error {
				factory.Close()
				return nil
			})
		}
		return factory, factory.Close
	}
}

func provideDefaultClient(maker Maker) (*mongo.Client, error) {
	return maker.Make("default")
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

// provideConfig exports the default mongo configuration.
func provideConfig() configOut {
	configs := []config.ExportedConfig{
		{
			Owner: "otmongo",
			Data: map[string]any{
				"mongo": map[string]struct {
					Uri string `json:"uri" yaml:"uri"`
				}{
					"default": {
						Uri: "mongodb://127.0.0.1:27017",
					},
				},
			},
			Comment: "The configuration of mongoDB",
		},
	}
	return configOut{Config: configs}
}
