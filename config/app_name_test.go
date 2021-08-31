package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppNameFromConf(t *testing.T) {
	t.Parallel()
	app := NewAppNameFromConf(WithAccessor(MapAdapter(map[string]interface{}{
		"name": "app",
	})))
	assert.Equal(t, "app", app.String())
}
