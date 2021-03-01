package kitkafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}
