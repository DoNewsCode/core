package ots3

import (
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
	"testing"
)

func TestNewUploadManagerFactory(t *testing.T) {
	factory := NewUploadManagerFactory(UploadManagerParam{
		In: dig.In{},
		Conf: config.MapAdapter{"s3": map[string]S3Config{
			"default":     {},
			"alternative": {},
		}},
		Tracer: nil,
	})
	assert.Len(t, factory.managers, 2)
}
