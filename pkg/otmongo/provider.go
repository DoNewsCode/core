package otmongo

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.uber.org/dig"
	"sync"

	"github.com/DoNewsCode/std/pkg/contract"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoParam struct {
	dig.In

	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

func Mongo(p MongoParam) (*mongo.Client, func(), error) {
	uri := p.Conf.String("mongo.default.uri")
	opts := options.Client()
	opts.ApplyURI(uri)
	if p.Tracer != nil {
		opts.Monitor = NewMonitor(p.Tracer)
	}
	client, err := mongo.Connect(context.Background(), opts)
	return client, func() {
		_ = client.Disconnect(context.Background())
	}, err
}

type MongoFactory struct {
	db map[string]*mongo.Client
}

func NewMongoFactory(p MongoParam) (*MongoFactory, func(), error) {
	var (
		uris map[string]struct {
			Uri string
		}
	)
	factory := &MongoFactory{
		db: make(map[string]*mongo.Client),
	}
	cleanup := func() {
		var wg sync.WaitGroup
		for i := range factory.db {
			wg.Add(1)
			go func(i string) {
				_ = factory.db[i].Disconnect(context.Background())
				wg.Done()
			}(i)
		}
		wg.Wait()
	}
	p.Conf.Unmarshal("mongo", &uris)
	for name, value := range uris {
		opts := options.Client()
		opts.ApplyURI(value.Uri)
		if p.Tracer != nil {
			opts.Monitor = NewMonitor(p.Tracer)
		}
		client, err := mongo.Connect(context.Background(), opts)
		if err != nil {
			return nil, cleanup, errors.Wrap(err, "failed to connect")
		}
		factory.db[name] = client
	}
	return factory, cleanup, nil
}
