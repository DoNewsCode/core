package ots3

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"

	"github.com/DoNewsCode/core/contract"
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

// Maker is an interface for *Factory. Used as a type hint for injection.
type Maker interface {
	Make(name string) (*Manager, error)
}

// in is the injection parameter for provideFactory.
type in struct {
	di.In

	Logger log.Logger
	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

// out is the di output of provideFactory.
type out struct {
	di.Out

	Factory Factory
	Maker   Maker
}

// Factory can be used to connect to multiple s3 servers.
type Factory struct {
	*di.Factory
}

// Make creates a s3 manager under the given name.
func (s Factory) Make(name string) (*Manager, error) {
	client, err := s.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*Manager), nil
}

// provideFactory creates *Factory and *ots3.Manager. It is a valid dependency for package core.
func provideFactory(p in) out {
	var (
		err       error
		s3configs map[string]S3Config
	)
	err = p.Conf.Unmarshal("s3", &s3configs)
	if err != nil {
		level.Warn(p.Logger).Log("err", err)
	}
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			ok   bool
			conf S3Config
		)
		if conf, ok = s3configs[name]; !ok {
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
	return out{
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
						AccessKey:    "",
						AccessSecret: "",
						Endpoint:     "",
						Region:       "",
						Bucket:       "",
						CdnUrl:       "",
					},
				}},
			Comment: "The s3 configuration",
		},
	}
	return configOut{Config: configs}
}
