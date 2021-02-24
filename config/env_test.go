package config

import (
	gotesting "testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv_String(t *gotesting.T) {
	assertion := assert.New(t)
	assertion.Equal("local", NewEnv("LOCAL").String())
	assertion.Equal("staging", NewEnv("STAGING").String())
	assertion.Equal("development", NewEnv("DEV").String())
	assertion.Equal("production", NewEnv("PRODUCTION").String())
	assertion.Equal("testing", NewEnv("TESTING").String())
}

func TestNewEnvFromConf(t *gotesting.T) {
	env := NewEnvFromConf(MapAdapter(map[string]interface{}{
		"env": "local",
	}))
	assert.Equal(t, "local", env.String())
}
