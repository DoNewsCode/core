package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	conf := provideDefaultConfig()
	assert.NotNil(t, conf)
}
