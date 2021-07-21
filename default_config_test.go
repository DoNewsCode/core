package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	conf := provideDefaultConfig()
	assert.NotNil(t, conf)

	for _, c := range conf {
		if c.Validate != nil {
			err := c.Validate(c.Data)
			assert.NoError(t, err)
		}
	}
}
