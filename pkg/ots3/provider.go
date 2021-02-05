package ots3

import (
	"fmt"
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

func ProvideUploadManager(conf contract.ConfigAccessor) *Manager {
	var s3config S3Config
	_ = conf.Unmarshal("s3.default", &s3config)
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
	)
}
