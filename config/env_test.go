package config

import (
	gotesting "testing"

	"github.com/stretchr/testify/assert"
)

var cases = []struct {
	name string
	env  string
	want Env
}{
	{"env-lower", "local", EnvLocal},
	{"env-upper", "LOCAL", EnvLocal},
	{"env-long", "production", EnvProduction},
	{"env-short", "prod", EnvProduction},
	{"env-alias", "online", EnvProduction},
	{"env-alias", "pre-prod", EnvStaging},
}

func TestEnv_String(t *gotesting.T) {
	t.Parallel()
	for _, c := range cases {
		t.Run(c.name, func(t *gotesting.T) {
			env := NewEnv(c.env)
			assert.Equal(t, c.want, env)
		})
	}
}

func TestNewEnvFromConf(t *gotesting.T) {
	t.Parallel()
	for _, c := range cases {
		t.Run(c.name, func(t *gotesting.T) {
			env := NewEnvFromConf(MapAdapter(map[string]any{
				"env": c.env,
			}))
			assert.Equal(t, c.want, env)
		})
	}
}
