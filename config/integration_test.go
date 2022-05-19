package config_test

import (
	"errors"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
)

func provideConfig() configOut {
	return configOut{
		Config: []config.ExportedConfig{{
			Validate: func(data map[string]any) error {
				return errors.New("bad config")
			},
		}},
	}
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

func Test_integration(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Errorf("test should panic. the config is not valid.")
	}()
	c := core.Default()
	c.Provide(di.Deps{
		provideConfig,
	})
	c.AddModuleFunc(config.New)
}
