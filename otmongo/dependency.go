package otmongo

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
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
func Providers() di.Deps {
	return []interface{}{provideMongoFactory, provideDefaultClient, provideConfig}
}

// MongoConfigInterceptor is an injection type hint that allows user to make last
// minute modification to mongo configuration. This is useful when some
// configuration cannot be easily expressed factoryIn a text form. For example, the
// options.ContextDialer.
type MongoConfigInterceptor func(name string, clientOptions *options.ClientOptions)

// factoryIn is the injection parameter for Provide.
type factoryIn struct {
	dig.In

	Logger      log.Logger
	Conf        contract.ConfigAccessor
	Interceptor MongoConfigInterceptor `optional:"true"`
	Tracer      opentracing.Tracer     `optional:"true"`
	Dispatcher  contract.Dispatcher    `optional:"true"`
}

// factoryOut is the result of Provide. The official mongo package doesn't
// provide a proper interface type. It is up to the users to define their own
// mongodb repository interface.
type factoryOut struct {
	dig.Out

	Factory Factory
	Maker   Maker
}

// Provide creates Factory and *mongo.Client. It is a valid dependency for
// package core.
func provideMongoFactory(p factoryIn) (factoryOut, func()) {
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			conf struct{ URI string }
		)
		if err := p.Conf.Unmarshal(fmt.Sprintf("mongo.%s", name), &conf); err != nil {
			return di.Pair{}, fmt.Errorf("mongo configuration %s not valid: %w", name, err)
		}
		if conf.URI == "" {
			conf.URI = "mongodb://127.0.0.1:27017"
		}

		opts := options.Client()
		opts.ApplyURI(conf.URI)
		if p.Tracer != nil {
			opts.Monitor = NewMonitor(p.Tracer)
		}
		if p.Interceptor != nil {
			p.Interceptor(name, opts)
		}
		client, err := mongo.Connect(context.Background(), opts)
		if err != nil {
			return di.Pair{}, err
		}
		return di.Pair{
			Conn: client,
			Closer: func() {
				_ = client.Disconnect(context.Background())
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
			Data: map[string]interface{}{
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
