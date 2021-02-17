package ots3

import (
	"fmt"
	"net/url"

	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"

	"github.com/DoNewsCode/std/pkg/contract"
)

// S3Config contains credentials of S3 server
type S3Config struct {
	AccessKey    string `json:"accessKey" yaml:"accessKey"`
	AccessSecret string `json:"accessSecret" yaml:"accessSecret"`
	Endpoint     string `json:"endpoint" yaml:"endpoint"`
	Region       string `json:"region" yaml:"region"`
	Bucket       string `json:"bucket" yaml:"bucket"`
	CdnUrl       string `json:"cdnUrl" yaml:"cdnUrl"`
}

// S3Maker is an interface for *S3Factory. Used as a type hint for injection.
type S3Maker interface {
	Make(name string) (*Manager, error)
}

// S3In is the injection parameter for ProvideManager.
type S3In struct {
	di.In

	Logger log.Logger
	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

// S3Out is the di output of ProvideManager.
type S3Out struct {
	di.Out
	di.Module

	Manager  *Manager
	Factory  *S3Factory
	Maker    S3Maker
	Uploader Uploader
}

// S3Factory can be used to connect to multiple s3 servers.
type S3Factory struct {
	*async.Factory
}

// Make creates a s3 manager under the given name.
func (s *S3Factory) Make(name string) (*Manager, error) {
	client, err := s.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*Manager), nil
}

// ProvideManager creates *S3Factory and *ots3.Manager. It is a valid dependency for package core.
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
	s3Factory := S3Factory{factory}
	manager, err := factory.Make("default")
	if err != nil {
		return S3Out{
			Manager:  nil,
			Uploader: nil,
			Factory:  &s3Factory,
			Maker:    &s3Factory,
		}
	}
	return S3Out{
		Manager:  manager.(*Manager),
		Uploader: manager.(*Manager),
		Factory:  &s3Factory,
		Maker:    &s3Factory,
	}
}

// ProvideConfig exports the default s3 configuration
func (m S3Out) ProvideConfig() []contract.ExportedConfig {
	return []contract.ExportedConfig{
		{
			Name: "s3",
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
}
