package ots3

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
)

/*
Providers returns a set of dependencies providers related to S3. It includes the s3
Manager, the Maker and exported configurations.
	Depends On:
		log.Logger
		contract.ConfigAccessor
		opentracing.Tracer `optional:"true"`
		contract.DIPopulator `optional:"true"`
	Provide:
		Factory
		Maker
		*Manager
		Uploader
*/
func Providers(optionFunc ...ProvidersOptionFunc) di.Deps {
	option := providersOption{
		ctor: newManager,
	}
	for _, f := range optionFunc {
		f(&option)
	}
	return di.Deps{provideFactory(&option), provideManager, provideConfig}
}

// Uploader models UploadService
type Uploader interface {
	// Upload the bytes from io.Reader with a given filename to a server, and returns the url and error.
	Upload(ctx context.Context, name string, reader io.Reader) (string, error)
}

// S3Config contains credentials of S3 server
type S3Config struct {
	AccessKey    string `json:"accessKey" yaml:"accessKey"`
	AccessSecret string `json:"accessSecret" yaml:"accessSecret"`
	Endpoint     string `json:"endpoint" yaml:"endpoint"`
	Region       string `json:"region" yaml:"region"`
	Bucket       string `json:"bucket" yaml:"bucket"`
	CdnUrl       string `json:"cdnUrl" yaml:"cdnUrl"`
}

// factoryIn is the injection parameter for provideFactory.
type factoryIn struct {
	di.In

	Logger     log.Logger
	Conf       contract.ConfigUnmarshaler
	Populator  contract.DIPopulator `optional:"true"`
	Dispatcher contract.Dispatcher  `optional:"true"`
}

// provideFactory creates *Factory and *ots3.Manager. It is a valid dependency for package core.
func provideFactory(option *providersOption) func(p factoryIn) Factory {
	if option.ctor == nil {
		option.ctor = newManager
	}
	return func(p factoryIn) Factory {
		factory := di.NewFactory(func(name string) (di.Pair, error) {

			var conf S3Config

			if err := p.Conf.Unmarshal(fmt.Sprintf("s3.%s", name), &conf); err != nil {
				if name != "default" {
					return di.Pair{}, fmt.Errorf("s3 configuration %s not found", name)
				}
				conf = S3Config{}
			}

			manager, err := option.ctor(ManagerArgs{
				Name:      name,
				Conf:      conf,
				Populator: p.Populator,
			})
			if err != nil {
				return di.Pair{}, fmt.Errorf("error constructing manager: %w", err)
			}
			return di.Pair{
				Closer: nil,
				Conn:   manager,
			}, nil
		})

		s3Factory := Factory{factory}
		s3Factory.SubscribeReloadEventFrom(p.Dispatcher)

		return s3Factory

	}
}

// ManagerArgs are arguments for constructing the s3 manager. When providing custom constructors, take this as input.
type ManagerArgs struct {
	Name      string
	Conf      S3Config
	Populator contract.DIPopulator
}

func newManager(args ManagerArgs) (*Manager, error) {
	var tracer opentracing.Tracer
	args.Populator.Populate(&tracer)
	manager := NewManager(
		args.Conf.AccessKey,
		args.Conf.AccessSecret,
		args.Conf.Endpoint,
		args.Conf.Region,
		args.Conf.Bucket,
		WithLocationFunc(func(location string) (uri string) {
			u, err := url.Parse(location)
			if err != nil {
				return location
			}
			return fmt.Sprintf(args.Conf.CdnUrl, u.Path[1:])
		}),
		WithTracer(tracer),
	)
	return manager, nil
}

type managerOut struct {
	di.Out

	Manager  *Manager
	Uploader Uploader
}

func provideManager(maker Maker) (managerOut, error) {
	manager, err := maker.Make("default")
	return managerOut{
		Manager:  manager,
		Uploader: manager,
	}, err
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

// provideConfig exports the default s3 configuration
func provideConfig() configOut {
	configs := []config.ExportedConfig{
		{
			Owner: "ots3",
			Data: map[string]interface{}{
				"s3": map[string]S3Config{
					"default": {
						AccessKey:    "http://127.0.0.1:9000",
						AccessSecret: "minioadmin",
						Endpoint:     "minioadmin",
						Region:       "asia",
						Bucket:       "mybucket",
						CdnUrl:       "",
					},
				}},
			Comment: "The s3 configuration",
		},
	}
	return configOut{Config: configs}
}
