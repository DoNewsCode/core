package ots3

import (
	"fmt"
	"github.com/DoNewsCode/std/pkg/async"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/dig"
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

type S3Param struct {
	dig.In

	Logger log.Logger
	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

func ProvideUploadManager(p S3Param) (*Manager, func(), error) {
	factory, _ := ProvideS3Factory(p)
	conn, err := factory.Make("default")
	return conn, func() {
		factory.CloseConn("default")
	}, err
}

type S3Factory struct {
	*async.Factory
}

func (s S3Factory) Make(name string) (*Manager, error) {
	client, err := s.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*Manager), nil
}

func ProvideS3Factory(p S3Param) (S3Factory, func()) {
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
		manager := NewManager(
			conf.AccessKey,
			conf.AccessSecret,
			conf.Endpoint,
			conf.Region,
			conf.Bucket,
			WithLocationFunc(func(location string) (uri string) {
				u, err := url.Parse(location)
				if err != nil {
					return location
				}
				return fmt.Sprintf(conf.CdnUrl, u.Path[1:])
			}),
			WithTracer(p.Tracer),
		)
		return async.Pair{
			Closer: nil,
			Conn:   manager,
		}, nil
	})
	return S3Factory{factory}, factory.Close
}
