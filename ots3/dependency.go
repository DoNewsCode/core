package ots3

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/internal"
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
	Provide:
		Factory
		Maker
		*Manager
		Uploader
*/
func Providers() []interface{} {
	return []interface{}{provideFactory, provideManager, provideConfig}
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
	Conf       contract.ConfigAccessor
	Tracer     opentracing.Tracer  `optional:"true"`
	Dispatcher contract.Dispatcher `optional:"true"`
}

// factoryOut is the di output of provideFactory.
type factoryOut struct {
	di.Out

	Factory Factory
	Maker   Maker
}

// provideFactory creates *Factory and *ots3.Manager. It is a valid dependency for package core.
func provideFactory(p factoryIn) factoryOut {
	factory := di.NewFactory(func(name string) (di.Pair, error) {

		var conf S3Config

		if err := p.Conf.Unmarshal(fmt.Sprintf("s3.%s", name), &conf); err != nil {
			if name != "default" {
				return di.Pair{}, fmt.Errorf("s3 configuration %s not found", name)
			}
			conf = S3Config{}
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
		return di.Pair{
			Closer: nil,
			Conn:   manager,
		}, nil
	})

	s3Factory := Factory{factory}
	s3Factory.SubscribeReloadEventFrom(p.Dispatcher)

	return factoryOut{
		Factory: s3Factory,
		Maker:   &s3Factory,
	}
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
						AccessKey:    envDefaultS3AccessKey,
						AccessSecret: envDefaultS3AccessSecret,
						Endpoint:     envDefaultS3Endpoint,
						Region:       envDefaultS3Region,
						Bucket:       envDefaultS3Bucket,
						CdnUrl:       "",
					},
				}},
			Comment: "The s3 configuration",
		},
	}
	return configOut{Config: configs}
}

var (
	envDefaultS3Endpoint, envDefaultS3EndpointIsSet = internal.GetDefaultAddrFromEnv("S3_ENDPOINT", "http://127.0.0.1:9000")
	envDefaultS3AccessKey, _                        = internal.GetDefaultAddrFromEnv("S3_ACCESSKEY", "minioadmin")
	envDefaultS3AccessSecret, _                     = internal.GetDefaultAddrFromEnv("S3_ACCESSSECRET", "minioadmin")
	envDefaultS3Region, _                           = internal.GetDefaultAddrFromEnv("S3_REGION", "asia")
	envDefaultS3Bucket, _                           = internal.GetDefaultAddrFromEnv("S3_BUCKET", "mybucket")
)
