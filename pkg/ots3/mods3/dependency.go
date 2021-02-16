package mods3

import (
	"fmt"
	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/DoNewsCode/std/pkg/ots3"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"
	"net/url"

	"github.com/DoNewsCode/std/pkg/contract"
)

type S3Config struct {
	AccessKey    string
	AccessSecret string
	Endpoint     string
	Region       string
	Bucket       string
	CdnUrl       string
}

type S3In struct {
	di.In

	Logger log.Logger
	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

type S3Out struct {
	di.Out

	Manager *ots3.Manager
	Factory *S3Factory
}

// S3Factory can be used to connect to multiple s3 servers.
type S3Factory struct {
	*async.Factory
}

// Make creates a s3 manager under the given name.
func (s *S3Factory) Make(name string) (*ots3.Manager, error) {
	client, err := s.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*ots3.Manager), nil
}

// ProvideManager creates S3Factory and *ots3.Manager. It is a valid dependency for package core.
func ProvideManager(p S3In) S3Out {
	var (
		err       error
		s3configs map[string]S3Config
	)
	err = p.Conf.Unmarshal("s3", &s3configs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok   bool
			conf S3Config
		)
		if conf, ok = s3configs[name]; !ok {
			return async.Pair{}, fmt.Errorf("s3 configuration %s not found", name)
		}
		manager := ots3.NewManager(
			conf.AccessKey,
			conf.AccessSecret,
			conf.Endpoint,
			conf.Region,
			conf.Bucket,
			ots3.WithLocationFunc(func(location string) (uri string) {
				u, err := url.Parse(location)
				if err != nil {
					return location
				}
				return fmt.Sprintf(conf.CdnUrl, u.Path[1:])
			}),
			ots3.WithTracer(p.Tracer),
		)
		return async.Pair{
			Closer: nil,
			Conn:   manager,
		}, nil
	})
	manager, err := factory.Make("default")
	if err != nil {
		return S3Out{
			Manager: nil,
			Factory: &S3Factory{factory},
		}
	}
	return S3Out{
		Manager: manager.(*ots3.Manager),
		Factory: &S3Factory{factory},
	}
}
