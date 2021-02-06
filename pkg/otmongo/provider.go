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

type MongoParam struct {
	dig.In

	Logger log.Logger
	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

func Mongo(p MongoParam) (*mongo.Client, func(), error) {
	factory, _ := ProvideMongoFactory(p)
	conn, err := factory.Make("default")
	return conn, func() {
		factory.CloseConn("default")
	}, err
}

type MongoFactory struct {
	*async.Factory
}

func (r MongoFactory) Make(name string) (*mongo.Client, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*mongo.Client), nil
}

func ProvideMongoFactory(p MongoParam) (MongoFactory, func()) {
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
			return async.Pair{}, fmt.Errorf("mongo configuration %s not valid", name)
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
	return MongoFactory{factory}, factory.Close
}
