package otmongo

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/dig"
)

// MongoIn is the injection parameter for ProvideMongo.
type MongoIn struct {
	dig.In

	Logger log.Logger
	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

// Maker models Factory
type Maker interface {
	Make(name string) (*mongo.Client, error)
}

// MongoOut is the result of ProvideMongo. The official mongo package doesn't
// provide a proper interface type. It is up to the users to define their own
// mongodb repository interface.
type MongoOut struct {
	dig.Out

	Factory        Factory
	Maker          Maker
	Client         *mongo.Client
	ExportedConfig []contract.ExportedConfig `group:"config,flatten"`
}

// ProvideMongo creates Factory and *mongo.Client. It is a valid dependency for
// package core.
func ProvideMongo(p MongoIn) (MongoOut, func()) {
	var err error
	var dbConfs map[string]struct{ Uri string }
	err = p.Conf.Unmarshal("mongo", &dbConfs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok   bool
			conf struct{ Uri string }
		)
		if conf, ok = dbConfs[name]; !ok {
			if name != "default" {
				return async.Pair{}, fmt.Errorf("mongo configuration %s not valid", name)
			}
			conf.Uri = "mongodb://127.0.0.1:27017"
		}
		opts := options.Client()
		opts.ApplyURI(conf.Uri)
		if p.Tracer != nil {
			opts.Monitor = NewMonitor(p.Tracer)
		}
		client, err := mongo.Connect(context.Background(), opts)
		if err != nil {
			return async.Pair{}, err
		}
		return async.Pair{
			Conn: client,
			Closer: func() {
				_ = client.Disconnect(context.Background())
			},
		}, nil
	})
	f := Factory{factory}
	client, _ := f.Make("default")
	return MongoOut{
		Factory:        f,
		Maker:          f,
		Client:         client,
		ExportedConfig: provideConfig(),
	}, factory.Close
}

// Factory is a *async.Factory that creates *mongo.Client using a specific
// configuration entry.
type Factory struct {
	*async.Factory
}

// Make creates *mongo.Client using a specific configuration entry.
func (r Factory) Make(name string) (*mongo.Client, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*mongo.Client), nil
}

// provideConfig exports the default mongo configuration.
func provideConfig() []contract.ExportedConfig {
	return []contract.ExportedConfig{
		{
			Owner: "otmongo",
			Data: map[string]interface{}{
				"mongo": map[string]struct {
					Uri string `json:"uri" yaml:"uri"`
				}{
					"default": {
						Uri: "",
					},
				},
			},
			Comment: "The configuration of mongoDB",
		},
	}
}
