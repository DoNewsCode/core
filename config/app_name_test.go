package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAppNameFromConf(t *testing.T) {
	t.Parallel()
	app := NewAppNameFromConf(MapAdapter(map[string]interface{}{
		"name": "app",
	}))
	assert.Equal(t, "app", app.String())
}
