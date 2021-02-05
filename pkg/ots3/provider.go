package ots3

import (
	"fmt"
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

type UploadManagerParam struct {
	dig.In

	Conf   contract.ConfigAccessor
	Tracer opentracing.Tracer `optional:"true"`
}

func ProvideUploadManager(param UploadManagerParam) *Manager {
	var s3config S3Config
	_ = param.Conf.Unmarshal("s3.default", &s3config)
	if param.Tracer == nil {
		param.Tracer = opentracing.NoopTracer{}
	}

	return NewManager(
		s3config.AccessKey,
		s3config.AccessSecret,
		s3config.Endpoint,
		s3config.Region,
		s3config.Bucket,
		WithLocationFunc(func(location string) (uri string) {
			u, err := url.Parse(location)
			if err != nil {
				return location
			}
			return fmt.Sprintf(s3config.CdnUrl, u.Path[1:])
		}),
		WithTracer(param.Tracer),
	)
}

type UploadManagerFactory struct {
	managers map[string]*Manager
}

func NewUploadManagerFactory(param UploadManagerParam) *UploadManagerFactory {
	var (
		s3config map[string]S3Config
		out      UploadManagerFactory
	)
	_ = param.Conf.Unmarshal("s3", &s3config)
	if param.Tracer == nil {
		param.Tracer = opentracing.NoopTracer{}
	}
	out.managers = make(map[string]*Manager)

	for name, value := range s3config {
		out.managers[name] = NewManager(
			value.AccessKey,
			value.AccessSecret,
			value.Endpoint,
			value.Region,
			value.Bucket,
			WithLocationFunc(func(location string) (uri string) {
				u, err := url.Parse(location)
				if err != nil {
					return location
				}
				return fmt.Sprintf(value.CdnUrl, u.Path[1:])
			}),
			WithTracer(param.Tracer),
		)
	}
	return &out
}

func (u *UploadManagerFactory) Connection(name string) *Manager {
	return u.managers[name]
}
