package ots3

import (
	"github.com/DoNewsCode/core/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewUploadManagerFactory(t *testing.T) {
	s3out := ProvideManager(S3In{
		Conf: config.MapAdapter{"s3": map[string]S3Config{
			"default":     {},
			"alternative": {},
		}},
		Tracer: nil,
	})
	def, err := s3out.Factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := s3out.Factory.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
}
